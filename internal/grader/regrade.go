package grader

import (
	"context"
	"errors"
	"fmt"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/jobmanager"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

type RegradeOptions struct {
	jobmanager.JobOptions

	Users []string `json:"users"`

	// Ensure all users have a new submission after this time.
	// If there is not a submission after this time,
	// the user's most recent submission will be regraded.
	// A value of zero will default to the current time.
	After timestamp.Timestamp `json:"after"`

	Options GradeOptions `json:"-"`

	Assignment *model.Assignment `json:"-"`

	// If true, do not swap the context to the background context when running.
	// By default (when this is false), the context will be swapped to the background context when !WaitForCompletion.
	// The swap is so that regrade does not get canceled when an HTTP request is complete.
	// Setting this true is useful for testing (as one round of regrade tests can be wrapped up).
	RetainOriginalContext bool `json:"-"`
}

func RegradeSubmissions(options RegradeOptions) (map[string]*model.SubmissionHistoryItem, int, error) {
	if !options.RetainOriginalContext && !options.WaitForCompletion {
		options.Context = context.Background()
	}

	if options.After == 0 {
		options.After = timestamp.Now()
	}

	// Lock based on the course to prevent multiple requests using up all the cores.
	lockKey := fmt.Sprintf("regrade-course-%s", options.Assignment.GetCourse().GetID())

	job := jobmanager.Job[string, *model.SubmissionHistoryItem]{
		JobOptions:              &options.JobOptions,
		LockKey:                 lockKey,
		PoolSize:                config.REGRADE_COURSE_POOL_SIZE.Get(),
		ReturnIncompleteResults: !options.WaitForCompletion,
		WorkItems:               options.Users,
		// TODO: Need a course get recent submission history that takes a list of users.
		RetrieveFunc: func(users []string) (map[string]*model.SubmissionHistoryItem, error) {
			return db.GetRecentSubmissionSurvey(options.Assignment, users)
		},
		WorkFunc: func(user string) (*model.SubmissionHistoryItem, error) {
			return runRegrade(options, user)
		},
		WorkItemKeyFunc: func(user string) string {
			return fmt.Sprintf("%s-%s", lockKey, user)
		},
	}

	err := job.Validate()
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to validate job: '%v'.", err)
	}

	output := job.Run()
	if output.Error != nil {
		return nil, 0, fmt.Errorf("Failed to run regrade job with '%d' work errors: '%v'.", len(output.WorkErrors), output.Error)
	}

	return output.ResultItems, len(output.RemainingItems), nil
}

func runRegrade(options RegradeOptions, user string) (*model.SubmissionHistoryItem, error) {
	previousResult, err := db.GetSubmissionContents(options.Assignment, user, "")
	if err != nil {
		return nil, fmt.Errorf("Failed to get most recent grading result for '%s': '%w'.", user, err)
	}

	if previousResult == nil {
		return nil, nil
	}

	dirName := fmt.Sprintf("regrade-%s-%s-%s-", options.Assignment.GetCourse().GetID(), options.Assignment.GetID(), user)
	tempDir, err := util.MkDirTemp(dirName)
	if err != nil {
		return nil, fmt.Errorf("Failed to create temp regrade dir: '%w'.", err)
	}
	defer util.RemoveDirent(tempDir)

	err = util.GzipBytesToDirectory(tempDir, previousResult.InputFilesGZip)
	if err != nil {
		return nil, fmt.Errorf("Failed to write submission input to a temp dir: '%v'.", err)
	}

	message := ""
	if previousResult.Info != nil {
		message = previousResult.Info.Message
		options.Options.ProxyTime = &previousResult.Info.GradingStartTime
	}

	gradingResult, reject, failureMessage, err := Grade(options.Context, options.Assignment, tempDir, user, message, options.Options)
	if err != nil {
		stdout := ""
		stderr := ""

		if (gradingResult != nil) && (gradingResult.HasTextOutput()) {
			stdout = gradingResult.Stdout
			stderr = gradingResult.Stderr
		}

		log.Warn("Regrade submission failed internally.", err, log.NewAttr("stdout", stdout), log.NewAttr("stderr", stderr))

		return nil, nil
	}

	if reject != nil {
		log.Debug("Regrade submission rejected.", log.NewAttr("reason", reject.String()))

		return nil, nil
	}

	if failureMessage != "" {
		log.Debug("Regrade submission got a soft error.", log.NewAttr("message", failureMessage))

		return nil, nil
	}

	var result *model.SubmissionHistoryItem = nil
	if gradingResult.Info != nil {
		result = (*gradingResult.Info).ToHistoryItem()
	}

	return result, nil
}
