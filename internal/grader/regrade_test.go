package grader

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/jobmanager"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestRegradeBase(test *testing.T) {
	docker.EnsureOrSkipForTest(test)

	defer db.ResetForTesting()

	// Note that computation of these paths are deferred until test time.
	studentGradingResults := map[string]*model.GradingResult{
		"1697406272": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406272")),
	}

	testCases := []struct {
		users             []string
		waitForCompletion bool
		results           []*model.SubmissionHistoryItem
	}{
		// User with submission, wait
		{
			[]string{"course-student@test.edulinq.org"},
			true,
			[]*model.SubmissionHistoryItem{
				studentGradingResults["1697406272"].Info.ToHistoryItem(),
			},
		},

		// Empty users, wait
		{
			[]string{},
			true,
			[]*model.SubmissionHistoryItem{},
		},

		// Empty submissions, wait
		{
			[]string{"course-admin@test.edulinq.org"},
			true,
			[]*model.SubmissionHistoryItem{
				nil,
			},
		},

		// User with submission, no wait
		{
			[]string{"course-student@test.edulinq.org"},
			false,
			[]*model.SubmissionHistoryItem{},
		},

		// Empty users, no wait
		{
			[]string{},
			false,
			[]*model.SubmissionHistoryItem{},
		},

		// Empty submission, no wait
		{
			[]string{"course-admin@test.edulinq.org"},
			false,
			[]*model.SubmissionHistoryItem{},
		},
	}

	assignment := db.MustGetTestAssignment()

	for i, testCase := range testCases {
		db.ResetForTesting()

		options := RegradeOptions{
			JobOptions: jobmanager.JobOptions{
				WaitForCompletion: testCase.waitForCompletion,
			},
			Options:    GetDefaultGradeOptions(),
			Users:      testCase.users,
			Assignment: assignment,
		}

		results, err := RegradeSubmissions(options)
		if err != nil {
			test.Errorf("Case %d: Failed to regrade submissions: '%v'.", i, err)
			continue
		}

		failed := CheckAndClearIDs(test, i, testCase.results, results)
		if failed {
			continue
		}

		if !reflect.DeepEqual(testCase.results, results) {
			test.Errorf("Case %d: Unexpected regrade result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.results), util.MustToJSONIndent(results))
			continue
		}
	}
}
