package proxy

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/grader"
	"github.com/edulinq/autograder/internal/jobmanager"
	"github.com/edulinq/autograder/internal/model"
)

type RegradeRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleGrader

	// Regrade the submissions for the following emails or roles.
	Emails []string `json:"emails"`

	// Wait for the entire regrade to complete and return all results.
	WaitForCompletion bool `json:"wait-for-completion"`
}

type RegradeResponse struct {
	Users   []string                       `json:"users"`
	Results []*model.SubmissionHistoryItem `json:"results"`
}

// Regrade the most recent submissions for users with the filtered role in the course.
func HandleRegrade(request *RegradeRequest) (*RegradeResponse, *core.APIError) {
	if len(request.Emails) == 0 {
		request.Emails = []string{"student"}
	}

	users, err := db.ResolveCourseUsers(request.Course, request.Emails)
	if err != nil {
		return nil, core.NewInternalError("-635", request, "Unable to resolve course users.")
	}

	gradeOptions := grader.GetDefaultGradeOptions()
	gradeOptions.ProxyUser = request.User.Email
	gradeOptions.CheckRejection = false

	regradeOptions := grader.RegradeOptions{
		JobOptions: jobmanager.JobOptions{
			WaitForCompletion: request.WaitForCompletion,
			Context:           request.Context,
		},
		Users:      users,
		Options:    gradeOptions,
		Assignment: request.Assignment,
	}

	results, err := grader.RegradeSubmissions(regradeOptions)
	if err != nil {
		return nil, core.NewInternalError("-636", request, "Unable to regrade subission contents.")
	}

	response := RegradeResponse{
		Users:   users,
		Results: results,
	}

	return &response, nil
}
