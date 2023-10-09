package grader

import (
    "fmt"
    "os"
    "path/filepath"
    "sync"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

var submissionLocks sync.Map;

type GradeOptions struct {
    UseFakeSubmissionsDir bool
    NoDocker bool
    LeaveTempDir bool
}

func GetDefaultGradeOptions() GradeOptions {
    return GradeOptions{
        UseFakeSubmissionsDir: config.NO_STORE.GetBool(),
        NoDocker: config.DOCKER_DISABLE.GetBool(),
        LeaveTempDir: config.DEBUG.GetBool(),
    };
}

// Grade with default options pulled from config.
func GradeDefault(assignment *model.Assignment, submissionPath string, user string, message string) (*model.GradedAssignment, *model.SubmissionSummary, string, error) {
    return Grade(assignment, submissionPath, user, message, GetDefaultGradeOptions());
}

// Grade with custom options.
func Grade(assignment *model.Assignment, submissionPath string, user string, message string, options GradeOptions) (*model.GradedAssignment, *model.SubmissionSummary, string, error) {
    gradingKey := fmt.Sprintf("%s::%s::%s", assignment.Course.ID, assignment.ID, user);

    // Get the existing mutex, or store (and fetch) a new one.
    val, _ := submissionLocks.LoadOrStore(gradingKey, &sync.Mutex{});
    lock := val.(*sync.Mutex)

    lock.Lock();
    defer lock.Unlock();

    submissionDir, submissionID, err := prepSubmissionDir(assignment, user, options);
    if (err != nil) {
        return nil, nil, "", fmt.Errorf("Failed to prepare submission dir for assignment '%s' and user '%s': '%w'.", assignment.FullID(), user, err);
    }

    gradingID := fmt.Sprintf("%s::%d", gradingKey, submissionID);

    // Copy the submission to the user's submission directory.
    submissionCopyDir := filepath.Join(submissionDir, common.GRADING_INPUT_DIRNAME);
    err = util.CopyDirent(submissionPath, submissionCopyDir, true);
    if (err != nil) {
        return nil, nil, "", fmt.Errorf("Failed to copy submission for assignment '%s' and user '%s': '%w'.", assignment.FullID(), user, err);
    }

    outputDir := filepath.Join(submissionDir, common.GRADING_OUTPUT_DIRNAME);
    os.MkdirAll(outputDir, 0755);

    var result *model.GradedAssignment;
    var output string;

    if (options.NoDocker) {
        result, output, err = RunNoDockerGrader(assignment, submissionPath, outputDir, options, gradingID);
    } else {
        result, output, err = RunDockerGrader(assignment, submissionPath, outputDir, options, gradingID);
    }

    if (err != nil) {
        return nil, nil, output, err;
    }

    summary := result.GetSummary(gradingID, message);
    summaryPath := filepath.Join(outputDir, common.GRADER_OUTPUT_SUMMARY_FILENAME);

    err = util.ToJSONFileIndent(summary, summaryPath);
    if (err != nil) {
        return nil, nil, output, fmt.Errorf("Failed to write submission summary for assignment '%s' and user '%s': '%w'.", assignment.FullID(), user, err);
    }

    return result, summary, output, nil;
}

func prepSubmissionDir(assignment *model.Assignment, user string, options GradeOptions) (string, int64, error) {
    var submissionDir string;
    var err error;
    var id int64;

    if (options.UseFakeSubmissionsDir) {
        tempSubmissionsDir, err := os.MkdirTemp("", "autograding-submissions-");
        if (err != nil) {
            return "", 0, fmt.Errorf("Could not create temp submissions dir: '%w'.", err);
        }

        submissionDir, id, err = assignment.PrepareSubmissionWithDir(user, tempSubmissionsDir);
        if (err != nil) {
            return "", 0, fmt.Errorf("Failed to prepare fake submission dir: '%w'.", err);
        }

        if (options.LeaveTempDir) {
            log.Info().Str("path", tempSubmissionsDir).Msg("Leaving behind temp submissions dir.");
        } else {
            defer os.RemoveAll(tempSubmissionsDir);
        }
    } else {
        submissionDir, id, err = assignment.PrepareSubmission(user);
        if (err != nil) {
            return "", 0, fmt.Errorf("Failed to prepare default submission dir: '%w'.", err);
        }
    }

    return submissionDir, id, nil;
}
