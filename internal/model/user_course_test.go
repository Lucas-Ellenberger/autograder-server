package model

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/util"
)

func TestUserCourseUserValidate(test *testing.T) {
	testCases := []struct {
		Email    string
		Name     *string
		Role     UserRole
		LMSID    *string
		Expected *CourseUser
	}{
		// Base
		{
			baseTestCourseUser.Email,
			baseTestCourseUser.Name,
			baseTestCourseUser.Role,
			baseTestCourseUser.LMSID,
			baseTestCourseUser,
		},

		// Email
		{
			" " + baseTestCourseUser.Email + " ",
			baseTestCourseUser.Name,
			baseTestCourseUser.Role,
			baseTestCourseUser.LMSID,
			baseTestCourseUser,
		},
		{
			"",
			baseTestCourseUser.Name,
			baseTestCourseUser.Role,
			baseTestCourseUser.LMSID,
			nil,
		},

		// Name
		{
			baseTestCourseUser.Email,
			util.StringPointer(" " + *baseTestCourseUser.Name + " "),
			baseTestCourseUser.Role,
			baseTestCourseUser.LMSID,
			baseTestCourseUser,
		},
		{
			baseTestCourseUser.Email,
			nil,
			baseTestCourseUser.Role,
			baseTestCourseUser.LMSID,
			setCourseUserName(baseTestCourseUser, nil),
		},
		{
			baseTestCourseUser.Email,
			util.StringPointer(""),
			baseTestCourseUser.Role,
			baseTestCourseUser.LMSID,
			setCourseUserName(baseTestCourseUser, util.StringPointer("")),
		},

		// Role
		{
			baseTestCourseUser.Email,
			baseTestCourseUser.Name,
			RoleUnknown,
			baseTestCourseUser.LMSID,
			nil,
		},

		// LMS ID
		{
			baseTestCourseUser.Email,
			baseTestCourseUser.Name,
			baseTestCourseUser.Role,
			util.StringPointer(" " + *baseTestCourseUser.LMSID + " "),
			baseTestCourseUser,
		},
		{
			baseTestCourseUser.Email,
			baseTestCourseUser.Name,
			baseTestCourseUser.Role,
			nil,
			setCourseUserLMSID(baseTestCourseUser, nil),
		},
		{
			baseTestCourseUser.Email,
			baseTestCourseUser.Name,
			baseTestCourseUser.Role,
			util.StringPointer(""),
			setCourseUserLMSID(baseTestCourseUser, util.StringPointer("")),
		},
	}

	for i, testCase := range testCases {
		user := &CourseUser{
			Email: testCase.Email,
			Name:  testCase.Name,
			Role:  testCase.Role,
			LMSID: testCase.LMSID,
		}

		err := user.Validate()
		if err != nil {
			if testCase.Expected == nil {
				// Expected failure.
				continue
			}

			test.Errorf("Case %d: User did not validate: '%v'.", i, err)
			continue
		}

		if testCase.Expected == nil {
			test.Errorf("Case %d: Expected failure did not happen: '%s'.", i, util.MustToJSONIndent(user))
			continue
		}

		if !reflect.DeepEqual(testCase.Expected, user) {
			test.Errorf("User not as expected. Expected: '%s', Actual: '%s'.",
				util.MustToJSONIndent(testCase.Expected), util.MustToJSONIndent(user))
			continue
		}
	}
}

func TestUserCourseUserName(test *testing.T) {
	testCases := []struct {
		BaseName     *string
		ResultName   string
		FallbackName string
	}{
		{util.StringPointer("foo"), "foo", "foo"},
		{util.StringPointer(""), "", "alice@test.com"},
		{nil, "", "alice@test.com"},
	}

	for i, testCase := range testCases {
		user := setCourseUserName(baseTestCourseUser, testCase.BaseName)
		err := user.Validate()
		if err != nil {
			test.Errorf("Case %d: Failed to validate user: '%v'.", i, err)
			continue
		}

		resultName := user.GetName(false)
		fallbackName := user.GetName(true)
		displayName := user.GetDisplayName()

		if testCase.ResultName != resultName {
			test.Errorf("Case %d: Result name not as expected. Expected: '%s', Actual: '%s'.", i, testCase.ResultName, resultName)
			continue
		}

		if testCase.FallbackName != fallbackName {
			test.Errorf("Case %d: Fallback name not as expected. Expected: '%s', Actual: '%s'.", i, testCase.FallbackName, fallbackName)
			continue
		}

		if testCase.FallbackName != displayName {
			test.Errorf("Case %d: Display name not as expected. Expected: '%s', Actual: '%s'.", i, testCase.FallbackName, displayName)
			continue
		}
	}
}

func setCourseUserName(user *CourseUser, name *string) *CourseUser {
	newUser := *user
	newUser.Name = name
	return &newUser
}

func setCourseUserRole(user *CourseUser, role UserRole) *CourseUser {
	newUser := *user
	newUser.Role = role
	return &newUser
}

func setCourseUserLMSID(user *CourseUser, lmsID *string) *CourseUser {
	newUser := *user
	newUser.LMSID = lmsID
	return &newUser
}

var baseTestCourseUser *CourseUser = &CourseUser{
	Email: "alice@test.com",
	Name:  util.StringPointer("Alice"),
	Role:  RoleStudent,
	LMSID: util.StringPointer("alice"),
}
