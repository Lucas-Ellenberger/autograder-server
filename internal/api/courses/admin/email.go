package admin

import (
	"errors"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/model"
)

type EmailRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleGrader

	email.Message

	DryRun bool `json:"dry-run"`
}

type EmailResponse struct {
	To  []model.CourseUserReferenceInput `json:"to"`
	CC  []model.CourseUserReferenceInput `json:"cc"`
	BCC []model.CourseUserReferenceInput `json:"bcc"`
}

// Send an email to course users.
func HandleEmail(request *EmailRequest) (*EmailResponse, *core.APIError) {
	if request.Subject == "" {
		return nil, core.NewBadRequestError("-627", request, "No email subject provided.")
	}

	var err error
	var errs error

	emailTo, err = db.ResolveCourseUsers(request.Course, request.To)
	errs = errors.Join(errs, err)

	emailCC, err = db.ResolveCourseUsers(request.Course, request.CC)
	errs = errors.Join(errs, err)

	emailBCC, err = db.ResolveCourseUsers(request.Course, request.BCC)
	errs = errors.Join(errs, err)

	if errs != nil {
		return nil, core.NewInternalError("-628", request, "Failed to resolve email recipients.").Err(errs)
	}

	if (len(emailTo) + len(emailCC) + len(emailBCC)) == 0 {
		return nil, core.NewBadRequestError("-629", request, "No email recipients provided.")
	}

	if !request.DryRun {
		err = email.SendFull(emailTo, emailCC, emailBCC, request.Subject, request.Body, request.HTML)
		if err != nil {
			return nil, core.NewInternalError("-630", request, "Failed to send email.")
		}
	}

	response := EmailResponse{
		To:  emailTo,
		CC:  emailCC,
		BCC: emailBCC,
	}

	return &response, nil
}
