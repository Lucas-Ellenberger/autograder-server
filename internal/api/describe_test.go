package api

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/util"
)

func TestDescribe(test *testing.T) {
	response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`describe`), nil, nil, "server-user")
	if !response.Success {
		test.Fatalf("Response is not a success when it should be: '%v'.", response)
	}

	var responseContent DescribeResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	expected := []string{
		"courses/admin/update",
		"courses/assignments/get",
		"courses/assignments/list",
		"courses/assignments/submissions/fetch/course/attemps",
		"courses/assignments/submissions/fetch/course/scores",
		"courses/assignments/submissions/fetch/user/attempt",
		"courses/assignments/submissions/fetch/user/attempts",
		"courses/assignments/submissions/fetch/user/history",
		"courses/assignments/submissions/fetch/user/peek",
		"courses/assignments/submissions/remove",
		"courses/assignments/submissions/submit",
		"courses/upsert/filespec",
		"courses/users/drop",
		"courses/users/enroll",
		"courses/users/get",
		"courses/users/list",
		"describe",
		"lms/upload/scores",
		"lms/user/get",
		"logs/query",
		"users/auth",
		"users/get",
		"users/list",
		"users/password/change",
		"users/password/reset",
		"users/remove",
		"users/tokens/create",
		"users/tokens/delete",
		"users/upsert",
	}

	if !reflect.DeepEqual(expected, responseContent.Endpoints) {
		test.Fatalf("Unexpected endpoints. Expected: '%s', actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(responseContent.Endpoints))
	}
}
