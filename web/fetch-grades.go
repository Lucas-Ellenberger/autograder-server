package web

import (
    "fmt"
    "net/http"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

var MIN_ROLE model.UserRole = model.Grader;

type FetchGradesRequest struct {
    model.BaseAPIRequest
    Assignment string `json:"assignment"`
}

type FetchGradesResponse struct {
    Summaries map[string]*model.SubmissionSummary `json:"result"`
}

func (this *FetchGradesRequest) String() string {
    return util.BaseString(this);
}

func NewFetchGradesRequest(request *http.Request) (*FetchGradesRequest, *model.APIResponse, error) {
    var apiRequest FetchGradesRequest;
    err := model.APIRequestFromPOST(&apiRequest, request);
    if (err != nil) {
        return nil, nil, err;
    }

    course, _, err := grader.VerifyCourseAssignment(apiRequest.Course, apiRequest.Assignment);
    if (err != nil) {
        return nil, nil, err;
    }

    ok, user, err := AuthAPIRequest(&apiRequest.BaseAPIRequest, course);
    if (err != nil) {
        return nil, nil, err;
    } else if (!ok) {
        return nil, model.NewResponse(http.StatusUnauthorized, "Failed to authenticate."), nil;
    }

    if ((user != nil) && (user.Role < MIN_ROLE)) {
        log.Debug().Str("user", user.Email).Msg("Authentication Failure: Insufficient Permissions.");
        return nil, model.NewResponse(http.StatusForbidden, "Insufficient Permissions."), nil;
    }

    return &apiRequest, nil, nil;
}

func (this *FetchGradesRequest) Close() error {
    return nil;
}

func (this *FetchGradesRequest) Clean() error {
    var err error;
    this.Assignment, err = model.ValidateID(this.Assignment);
    if (err != nil) {
        return fmt.Errorf("Could not clean FetchGradesRequest assignment ID ('%s'): '%w'.", this.Assignment, err);
    }

    return nil;
}

func handleFetchGrades(request *FetchGradesRequest) (int, any, error) {
    assignment := grader.GetAssignment(request.Course, request.Assignment);
    if (assignment == nil) {
        return http.StatusBadRequest, fmt.Sprintf("Could not find assignment ('%s') for course ('%s').", request.Assignment, request.Course,), nil;
    }

    users, err := assignment.Course.GetUsers();
    if (err != nil) {
        return 0, nil, fmt.Errorf("Failed to get users for course: '%w'.", err);
    }

    paths, err := assignment.GetAllRecentSubmissionSummaries(users);
    if (err != nil) {
        return 0, nil, fmt.Errorf("Failed to get submission summaries: '%w'.", err);
    }

    response := FetchGradesResponse{
        Summaries: make(map[string]*model.SubmissionSummary, len(paths)),
    };

    for email, path := range paths {
        if (path == "") {
            response.Summaries[email] = nil;
            continue;
        }

        summary := model.SubmissionSummary{};
        err = util.JSONFromFile(path, &summary);
        if (err != nil) {
            return 0, nil, fmt.Errorf("Failed to deserialize submission summary '%s': '%w'.", path, err);
        }

        response.Summaries[email] = &summary;
    }

    return 0, response, nil;
}