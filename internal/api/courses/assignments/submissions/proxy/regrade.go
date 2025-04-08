package proxy

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/grader"
	"github.com/edulinq/autograder/internal/model"
)

type RegradeRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleGrader

	// Filter results to only users with this role.
	FilterRole model.CourseUserRole `json:"filter-role"`
}

type RegradeResponse struct {
	Complete bool                                    `json:"complete"`
	Options  grader.RegradeOptions                   `json:"options"`
	Results  map[string]*model.SubmissionHistoryItem `json:"results"`
}

// Regrade the most recent submissions for users with the filtered role in the course.
func HandleRegrade(request *RegradeRequest) (*RegradeResponse, *core.APIError) {
	role := model.GetCourseUserRoleString(request.FilterRole)
	users, err := db.ResolveCourseUsers(request.Course, []string{role})
	if err != nil {
		return nil, core.NewInternalError("-635", request, "Unable to resolve course users.")
	}

	gradeOptions := grader.GetDefaultGradeOptions()
	gradeOptions.ProxyUser = request.User.Email
	gradeOptions.CheckRejection = false

	regradeOptions := grader.RegradeOptions{
		Options:    gradeOptions,
		Context:    request.Context,
		Assignment: request.Assignment,
		Users:      users,
	}

	results, pendingCount, err := grader.RegradeSubmissions(regradeOptions)
	if err != nil {
		return nil, core.NewInternalError("-636", request, "Unable to regrade subission contents.")
	}

	response := RegradeResponse{
		Complete: (pendingCount == 0),
		Options:  regradeOptions,
		Results:  results,
	}

	return &response, nil
}
