package grader

import (
	"fmt"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/jobmanager"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

type RegradeOptions struct {
	jobmanager.JobOptions

	Users []string `json:"users"`

	Options GradeOptions `json:"-"`

	Assignment *model.Assignment `json:"-"`
}

func RegradeSubmissions(options RegradeOptions) ([]*model.SubmissionHistoryItem, error) {
	options.LockKey = fmt.Sprintf("regrade-course-%s", options.Assignment.GetCourse().GetID())

	options.PoolSize = config.REGRADE_COURSE_POOL_SIZE.Get()

	err := options.JobOptions.Validate()
	if err != nil {
		return nil, fmt.Errorf("Failed to validate job options: '%v'.", err)
	}

	job := jobmanager.Job[string, *model.SubmissionHistoryItem]{
		JobOptions: options.JobOptions,
		WorkItems:  options.Users,
		WorkFunc: func(user string) (*model.SubmissionHistoryItem, int64, error) {
			return runSingleRegrade(options, user)
		},
	}

	output, err := job.Run()
	if err != nil {
		return nil, fmt.Errorf("Failed to run regrade job: '%v'.", err)
	}

	return output.ResultItems, nil
}

func runSingleRegrade(options RegradeOptions, user string) (*model.SubmissionHistoryItem, int64, error) {
	previousResult, err := db.GetSubmissionContents(options.Assignment, user, "")
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to get most recent grading result for '%s': '%w'.", user, err)
	}

	if previousResult == nil {
		return nil, 0, nil
	}

	dirName := fmt.Sprintf("regrade-%s-%s-%s-", options.Assignment.GetCourse().GetID(), options.Assignment.GetID(), user)
	tempDir, err := util.MkDirTemp(dirName)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to create temp regrade dir: '%w'.", err)
	}
	defer util.RemoveDirent(tempDir)

	err = util.GzipBytesToDirectory(tempDir, previousResult.InputFilesGZip)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to write submission input to a temp dir: '%v'.", err)
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

		return nil, 0, nil
	}

	if reject != nil {
		log.Debug("Regrade submission rejected.", log.NewAttr("reason", reject.String()))

		return nil, 0, nil
	}

	if failureMessage != "" {
		log.Debug("Regrade submission got a soft error.", log.NewAttr("message", failureMessage))

		return nil, 0, nil
	}

	var result *model.SubmissionHistoryItem = nil
	if gradingResult.Info != nil {
		result = (*gradingResult.Info).ToHistoryItem()
	}

	return result, 0, nil
}
