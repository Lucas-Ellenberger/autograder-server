package main

import (
	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

var args struct {
	config.ConfigArgs
	Course  string                           `help:"Optional Course ID. Only required when roles or * (all course users) are in the recipients." arg:"" optional:""`
	To      []model.CourseUserReferenceInput `help:"Email recipents (to)." required:""`
	CC      []string                         `help:"Email recipents (cc)." optional:""`
	BCC     []string                         `help:"Email recipents (bcc)." optional:""`
	Subject string                           `help:"Email subject." required:""`
	Body    string                           `help:"Email body." required:""`
}

func main() {
	kong.Parse(&args,
		kong.Description("Send an email."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	db.MustOpen()
	defer db.MustClose()

	var emailTo []string = nil
	if args.Course != "" {
		course := db.MustGetCourse(args.Course)

		emailTo, err = db.ResolveCourseUsers(course, args.To)
		if err != nil {
			log.Fatal("Failed to resolve users.", err, course)
		}
	} else {
		emailTo = model.CourseUserReferenceInputToStrings(args.To)
	}

	err = email.SendFull(emailTo, args.CC, args.BCC, args.Subject, args.Body, false)
	if err != nil {
		log.Fatal("Could not send email.", err)
	}

	err = email.Close()
	if err != nil {
		log.Fatal("Failed to close SMTP connection.", err)
	}
}
