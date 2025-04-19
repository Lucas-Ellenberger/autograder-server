package proxy

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/grader"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestRegradeBase(test *testing.T) {
	docker.EnsureOrSkipForTest(test)

	db.ResetForTesting()
	defer db.ResetForTesting()

	// Note that computation of these paths are deferred until test time.
	studentGradingResults := map[string]*model.GradingResult{
		"1697406272": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406272")),
	}

	testCases := []struct {
		emails            []string
		proxyUser         string
		waitForCompletion bool
		expectedLocator   string
		expected          RegradeResponse
	}{
		// Valid Regrade Submissions

		// Student, Wait For Completion
		{
			[]string{"student"},
			"course-grader",
			true,
			"",
			RegradeResponse{
				Users: []string{"course-student@test.edulinq.org"},
				Results: []*model.SubmissionHistoryItem{
					studentGradingResults["1697406272"].Info.ToHistoryItem(),
				},
			},
		},

		// Target Student, Wait For Completion
		{
			[]string{"course-student@test.edulinq.org"},
			"course-grader",
			true,
			"",
			RegradeResponse{
				Users: []string{"course-student@test.edulinq.org"},
				Results: []*model.SubmissionHistoryItem{
					studentGradingResults["1697406272"].Info.ToHistoryItem(),
				},
			},
		},

		// Admin, Wait For Completion
		{
			[]string{"admin"},
			"course-grader",
			true,
			"",
			RegradeResponse{
				Users: []string{"course-admin@test.edulinq.org"},
				Results: []*model.SubmissionHistoryItem{
					nil,
				},
			},
		},

		// All, Wait For Completion
		{
			[]string{"*"},
			"course-grader",
			true,
			"",
			RegradeResponse{
				Users: []string{
					"course-admin@test.edulinq.org",
					"course-grader@test.edulinq.org",
					"course-other@test.edulinq.org",
					"course-owner@test.edulinq.org",
					"course-student@test.edulinq.org",
				},
				Results: []*model.SubmissionHistoryItem{
					nil,
					nil,
					nil,
					nil,
					studentGradingResults["1697406272"].Info.ToHistoryItem(),
				},
			},
		},

		// Student, No Wait
		{
			[]string{"student"},
			"course-grader",
			false,
			"",
			RegradeResponse{
				Users:   []string{"course-student@test.edulinq.org"},
				Results: []*model.SubmissionHistoryItem{},
			},
		},

		// Grader, No Wait
		{
			[]string{"grader"},
			"course-grader",
			false,
			"",
			RegradeResponse{
				Users:   []string{"course-grader@test.edulinq.org"},
				Results: []*model.SubmissionHistoryItem{},
			},
		},

		// All, No Wait
		{
			[]string{"*"},
			"course-grader",
			false,
			"",
			RegradeResponse{
				Users: []string{
					"course-admin@test.edulinq.org",
					"course-grader@test.edulinq.org",
					"course-other@test.edulinq.org",
					"course-owner@test.edulinq.org",
					"course-student@test.edulinq.org",
				},
				Results: []*model.SubmissionHistoryItem{},
			},
		},

		// Unknown Users, Wait
		{
			[]string{"ZZZ"},
			"course-admin",
			true,
			"",
			RegradeResponse{
				Users:   []string{},
				Results: []*model.SubmissionHistoryItem{},
			},
		},
		{
			[]string{""},
			"course-admin",
			true,
			"",
			RegradeResponse{
				Users:   []string{},
				Results: []*model.SubmissionHistoryItem{},
			},
		},

		// Unknown Users, no wait
		{
			[]string{"ZZZ"},
			"course-admin",
			false,
			"",
			RegradeResponse{
				Users:   []string{},
				Results: []*model.SubmissionHistoryItem{},
			},
		},

		// Invalid Regrade Submissions

		// Perm Errors
		{
			[]string{"student"},
			"course-student",
			false,
			"-020",
			RegradeResponse{},
		},
		{
			[]string{"student"},
			"course-other",
			true,
			"-020",
			RegradeResponse{},
		},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"emails":              testCase.emails,
			"wait-for-completion": testCase.waitForCompletion,
		}

		response := core.SendTestAPIRequestFull(test, `courses/assignments/submissions/proxy/regrade`, fields, nil, testCase.proxyUser)
		if !response.Success {
			if testCase.expectedLocator != "" {
				if response.Locator != testCase.expectedLocator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
						i, testCase.expectedLocator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		if testCase.expectedLocator != "" {
			test.Errorf("Case %d: Did not get an expected error.", i)
			continue
		}

		var responseContent RegradeResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		failed := grader.CheckAndClearIDs(test, i, testCase.expected.Results, responseContent.Results)
		if failed {
			continue
		}

		if !reflect.DeepEqual(testCase.expected, responseContent) {
			test.Errorf("Case %d: Unexpected regrade result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(responseContent))
			continue
		}
	}
}
