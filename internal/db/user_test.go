package db

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func (this *DBTests) DBTestUserGetServerUsersBase(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	expected := mustLoadTestServerUsers()

	users, err := GetServerUsers()
	if err != nil {
		test.Fatalf("Could not get server users: '%v'.", err)
	}

	if len(users) == 0 {
		test.Fatalf("Found no server users.")
	}

	if !reflect.DeepEqual(expected, users) {
		test.Fatalf("Server users do not match. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(users))
	}
}

func (this *DBTests) DBTestUserGetServerUsersEmpty(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	MustClear()

	users, err := GetServerUsers()
	if err != nil {
		test.Fatalf("Could not get server users: '%v'.", err)
	}

	if len(users) != 0 {
		test.Fatalf("Found server users when there should have been none: '%s'.", util.MustToJSONIndent(users))
	}
}

func (this *DBTests) DBTestUserGetCourseUsersBase(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	course := MustGetTestCourse()

	serverUsers := mustLoadTestServerUsers()
	expected := convertToCourseUsers(test, course, serverUsers)

	testCourseUsers(test, course, expected)
}

func (this *DBTests) DBTestUserGetCourseUsersEmpty(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	course := MustGetCourse("course-languages")

	users, err := GetCourseUsers(course)
	if err != nil {
		test.Fatalf("Could not get initial course users: '%v'.", err)
	}

	if len(users) == 0 {
		test.Fatalf("Could not find any users when there should be some.")
	}

	// Clear the db (users) and re-add the courses without server-level users..
	MustClear()
	MustAddCourses()

	course = MustGetCourse("course-languages")

	users, err = GetCourseUsers(course)
	if err != nil {
		test.Fatalf("Could not get course users: '%v'.", err)
	}

	if len(users) != 0 {
		test.Fatalf("Found course users when there should have been none: '%s'.", util.MustToJSONIndent(users))
	}
}

func (this *DBTests) DBTestUserGetServerUserBase(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	email := "student@test.com"
	expected := mustLoadTestServerUsers()[email]

	user, err := GetServerUser(email, true)
	if err != nil {
		test.Fatalf("Could not get server user ('%s'): '%v'.", email, err)
	}

	if user == nil {
		test.Fatalf("Got nil server user ('%s').", email)
	}

	if !reflect.DeepEqual(expected, user) {
		test.Fatalf("Server user does not match. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(user))
	}
}

func (this *DBTests) DBTestUserGetServerUserNoTokens(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	email := "student@test.com"
	expected := mustLoadTestServerUsers()[email]
	expected.Tokens = nil

	user, err := GetServerUser(email, false)
	if err != nil {
		test.Fatalf("Could not get server user ('%s'): '%v'.", email, err)
	}

	if user == nil {
		test.Fatalf("Got nil server user ('%s').", email)
	}

	if !reflect.DeepEqual(expected, user) {
		test.Fatalf("Server user does not match. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(user))
	}
}

func (this *DBTests) DBTestUserGetServerUserMissing(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	email := "ZZZ"

	user, err := GetServerUser(email, false)
	if err != nil {
		test.Fatalf("Could not get server user ('%s'): '%v'.", email, err)
	}

	if user != nil {
		test.Fatalf("Got a non-nil server user ('%s') when there should be no user.", email)
	}
}

func (this *DBTests) DBTestUserGetCourseUserBase(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	course := MustGetTestCourse()
	email := "student@test.com"

	expected, err := mustLoadTestServerUsers()[email].ToCourseUser(course.ID)
	if err != nil {
		test.Fatalf("Could not get expected course user ('%s'): '%v'.", email, err)
	}

	user, err := GetCourseUser(course, email)
	if err != nil {
		test.Fatalf("Could not get course user ('%s'): '%v'.", email, err)
	}

	if user == nil {
		test.Fatalf("Got nil course user ('%s').", email)
	}

	if !reflect.DeepEqual(expected, user) {
		test.Fatalf("Course user does not match. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(user))
	}
}

func (this *DBTests) DBTestUserGetCourseUserMissing(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	course := MustGetTestCourse()
	email := "ZZZ"

	user, err := GetCourseUser(course, email)
	if err != nil {
		test.Fatalf("Could not get course user ('%s'): '%v'.", email, err)
	}

	if user != nil {
		test.Fatalf("Got a non-nil course user ('%s') when there should be no user.", email)
	}
}

func (this *DBTests) DBTestUserGetCourseUserNotEnrolled(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	course := MustGetTestCourse()
	email := "student@test.com"

	_, _, err := RemoveUserFromCourse(course, email)
	if err != nil {
		test.Fatalf("Failed to remove user from course: '%v'.", err)
	}

	user, err := GetCourseUser(course, email)
	if err != nil {
		test.Fatalf("Could not get course user ('%s'): '%v'.", email, err)
	}

	if user != nil {
		test.Fatalf("Got a non-nil course user ('%s') when there should be no user.", email)
	}
}

func (this *DBTests) DBTestUserUpsertUserInsert(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	email := "new@test.com"
	name := "new"

	expected := &model.ServerUser{
		Email: email,
		Name:  &name,
	}

	err := UpsertUser(expected)
	if err != nil {
		test.Fatalf("Could not upsert user '%s': '%v'.", email, err)
	}

	newUser, err := GetServerUser(email, true)
	if err != nil {
		test.Fatalf("Could not get (new) server user ('%s'): '%v'.", email, err)
	}

	if newUser == nil {
		test.Fatalf("Got nil (new) server user ('%s').", email)
	}

	if !reflect.DeepEqual(expected, newUser) {
		test.Fatalf("Server user does not match. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(newUser))
	}
}

func (this *DBTests) DBTestUserUpsertUserUpdate(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	email := "student@test.com"
	expected := mustLoadTestServerUsers()[email]

	newExpectedName := "Test Name"
	expected.Name = &newExpectedName

	user, err := GetServerUser(email, true)
	if err != nil {
		test.Fatalf("Could not get server user ('%s'): '%v'.", email, err)
	}

	if user == nil {
		test.Fatalf("Got nil server user ('%s').", email)
	}

	newActualName := "Test Name"
	user.Name = &newActualName

	// Remove any additive components.
	user.Tokens = make([]*model.Token, 0)

	err = UpsertUser(user)
	if err != nil {
		test.Fatalf("Could not upsert user '%s': '%v'.", email, err)
	}

	newUser, err := GetServerUser(email, true)
	if err != nil {
		test.Fatalf("Could not get (new) server user ('%s'): '%v'.", email, err)
	}

	if newUser == nil {
		test.Fatalf("Got nil (new) server user ('%s').", email)
	}

	if !reflect.DeepEqual(expected, newUser) {
		test.Fatalf("Server user does not match. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(newUser))
	}
}

func (this *DBTests) DBTestUserUpsertUserEmptyUpdate(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	email := "student@test.com"
	expected := mustLoadTestServerUsers()[email]

	user, err := GetServerUser(email, true)
	if err != nil {
		test.Fatalf("Could not get server user ('%s'): '%v'.", email, err)
	}

	if user == nil {
		test.Fatalf("Got nil server user ('%s').", email)
	}

	// Remove any additive components.
	user.Tokens = make([]*model.Token, 0)

	err = UpsertUser(user)
	if err != nil {
		test.Fatalf("Could not upsert user '%s': '%v'.", email, err)
	}

	newUser, err := GetServerUser(email, true)
	if err != nil {
		test.Fatalf("Could not get (new) server user ('%s'): '%v'.", email, err)
	}

	if newUser == nil {
		test.Fatalf("Got nil (new) server user ('%s').", email)
	}

	if !reflect.DeepEqual(expected, newUser) {
		test.Fatalf("Server user does not match. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(newUser))
	}
}

func (this *DBTests) DBTestUserUpsertCourseUserInsert(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	course := MustGetTestCourse()
	email := "new@test.com"
	name := "new"

	expected := &model.CourseUser{
		Email: email,
		Name:  &name,
		Role:  model.RoleStudent,
	}

	err := UpsertCourseUser(course, expected)
	if err != nil {
		test.Fatalf("Could not upsert user '%s': '%v'.", email, err)
	}

	newUser, err := GetCourseUser(course, email)
	if err != nil {
		test.Fatalf("Could not get (new) course user ('%s'): '%v'.", email, err)
	}

	if newUser == nil {
		test.Fatalf("Got nil (new) course user ('%s').", email)
	}

	if !reflect.DeepEqual(expected, newUser) {
		test.Fatalf("Course user does not match. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(newUser))
	}
}

func (this *DBTests) DBTestUserDeleteUserBase(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	email := "student@test.com"

	exists, err := DeleteUser(email)
	if err != nil {
		test.Fatalf("Could not delete user '%s': '%v'.", email, err)
	}

	if !exists {
		test.Fatalf("Told that user ('%s') did not exist, when it should have.", email)
	}

	user, err := GetServerUser(email, true)
	if err != nil {
		test.Fatalf("Could not get (new) server user ('%s'): '%v'.", email, err)
	}

	if user != nil {
		test.Fatalf("Got (new) server user ('%s') when it should have been deleted.", email)
	}
}

func (this *DBTests) DBTestUserDeleteUserMissing(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	email := "ZZZ"

	exists, err := DeleteUser(email)
	if err != nil {
		test.Fatalf("Could not delete user '%s': '%v'.", email, err)
	}

	if exists {
		test.Fatalf("Told that user ('%s') exists, when it should not.", email)
	}

	user, err := GetServerUser(email, true)
	if err != nil {
		test.Fatalf("Could not get (new) server user ('%s'): '%v'.", email, err)
	}

	if user != nil {
		test.Fatalf("Got (new) server user ('%s') when it should have been deleted.", email)
	}
}

func (this *DBTests) DBTestUserRemoveUserFromCourseBase(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	course := MustGetTestCourse()

	testCases := []struct {
		email    string
		exists   bool
		enrolled bool
	}{
		{"student@test.com", true, true},
		// Note that we will not reset between test cases.
		{"student@test.com", true, false},
		{"ZZZ", false, false},
	}

	for i, testCase := range testCases {
		exists, enrolled, err := RemoveUserFromCourse(course, testCase.email)
		if err != nil {
			test.Errorf("Case %d: Failed to remove user ('%s') from course: '%v'.", i, testCase.email, err)
			continue
		}

		if testCase.exists != exists {
			test.Errorf("Case %d: Unexpected exists. Expected: %v, Actual: %v.", i, testCase.exists, exists)
			continue
		}

		if testCase.enrolled != enrolled {
			test.Errorf("Case %d: Unexpected enrolled. Expected: %v, Actual: %v.", i, testCase.enrolled, enrolled)
			continue
		}

		user, err := GetCourseUser(course, testCase.email)
		if err != nil {
			test.Fatalf("Could not get course user ('%s'): '%v'.", testCase.email, err)
		}

		if user != nil {
			test.Fatalf("Got a non-nil course user ('%s') when there should be no user enrolled.", testCase.email)
		}
	}
}

func testCourseUsers(test *testing.T, course *model.Course, expected map[string]*model.CourseUser) {
	users, err := GetCourseUsers(course)
	if err != nil {
		test.Fatalf("Could not get course users: '%v'.", err)
	}

	if len(users) == 0 {
		test.Fatalf("Found no course users.")
	}

	if !reflect.DeepEqual(expected, users) {
		test.Fatalf("Course users do not match. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(users))
	}
}

func mustLoadTestServerUsers() map[string]*model.ServerUser {
	path := filepath.Join(config.GetCourseImportDir(), "testdata", model.USERS_FILENAME)

	users, err := model.LoadServerUsersFile(path)
	if err != nil {
		log.Fatal("Could not open test users file.", err, log.NewAttr("path", path))
	}

	return users
}

func convertToCourseUsers(test *testing.T, course *model.Course, serverUsers map[string]*model.ServerUser) map[string]*model.CourseUser {
	courseUsers := make(map[string]*model.CourseUser, len(serverUsers))
	for email, serverUser := range serverUsers {
		courseUser, err := serverUser.ToCourseUser(course.ID)
		if err != nil {
			test.Fatalf("Could not convert server user to course user: '%v'.", err)
		}

		if courseUser != nil {
			courseUsers[email] = courseUser
		}
	}

	return courseUsers
}
