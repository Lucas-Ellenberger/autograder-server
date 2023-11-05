package lmsusers

import (
    "path/filepath"
    "slices"
    "testing"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/email"
    "github.com/eriq-augustine/autograder/lms"
    lmstest "github.com/eriq-augustine/autograder/lms/adapter/test"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

type SyncLMSTestCase struct {
    dryRun bool
    sendEmails bool
    syncAttributes bool
    syncAdd bool
    syncDel bool
}

func TestCourseSyncLMSUsers(test *testing.T) {
    course := db.MustGetTestCourse();

    defer resetAdapter(course);

    testLMSAdapter := course.GetLMSAdapter().Adapter.(*lmstest.TestLMSAdapter);
    testLMSAdapter.SetUsersModifier(testingUsers);

    // Quiet the output a bit.
    oldLevel := config.GetLoggingLevel();
    config.SetLogLevelFatal();
    defer config.SetLoggingLevel(oldLevel);

    for i, testCase := range getSyncLMSTestCases() {
        course.GetLMSAdapter().SyncUserAttributes = testCase.syncAttributes;
        course.GetLMSAdapter().SyncAddUsers = testCase.syncAdd;
        course.GetLMSAdapter().SyncRemoveUsers = testCase.syncDel;

        email.ClearTestMessages();

        result, err := SyncLMSUsers(course, testCase.dryRun, testCase.sendEmails);
        if (err != nil) {
            test.Errorf("Case %d (%+v): User sync failed: '%v'.", i, testCase, err);
            continue;
        }

        testMessages := email.GetTestMessages();

        // LMS syncs cannot skip users.
        if (len(result.Skip) != 0) {
            test.Errorf("Case %d (%+v): Skipped users is not empty.", i, testCase);
            continue;
        }

        // There will always be mod users, since LMS IDs are always synced.
        // But when the option is on, additional attriutes will be synced.
        currentModUsers := modUsers;
        if (testCase.syncAttributes) {
            currentModUsers = modAllUsers;
        }

        if (!usr.UsersPointerEqual(currentModUsers, result.Mod)) {
            test.Errorf("Case %d (%+v): Unexpected mod users. Expected: '%s', actual: '%s'.",
                    i, testCase, util.MustToJSON(currentModUsers), util.MustToJSON(result.Mod));
            continue;
        }

        if (testCase.syncAdd) {
            if (!usr.UsersPointerEqual(addUsers, result.Add)) {
                test.Errorf("Case %d (%+v): Unexpected add users. Expected: '%s', actual: '%s'.",
                        i, testCase, util.MustToJSON(addUsers), util.MustToJSON(result.Add));
                continue;
            }

            if (len(result.Add) != len(result.ClearTextPasswords)) {
                test.Errorf("Case %d (%+v): Number of cleartext passwords (%d) does not match number of add users (%d).",
                        i, testCase, len(result.ClearTextPasswords), len(result.Add));
                continue;
            }

            for _, user := range addUsers {
                _, ok := result.ClearTextPasswords[user.Email];
                if (!ok) {
                    test.Errorf("Case %d (%+v): Add user '%s' does not have a cleartext password.", i, testCase, user.Email);
                    continue;
                }
            }

            if (testCase.dryRun || !testCase.sendEmails) {
                if (len(testMessages) != 0) {
                    test.Errorf("Case %d (%+v): User additions were enabled on a no-email/dry run, but %d new emails were found.", i, testCase, len(testMessages));
                    continue;
                }
            } else {
                if (!email.ShallowSliceEqual(addEmails, testMessages)) {
                    test.Errorf("Case %d (%+v): Unexpected add emails. Expected: '%s', actual: '%s'.",
                            i, testCase, util.MustToJSON(addEmails), util.MustToJSON(testMessages));
                    continue;
                }
            }
        } else {
            if (len(result.Add) != 0) {
                test.Errorf("Case %d (%+v): User additions were disabled, but %d new users were found.", i, testCase, len(result.Add));
                continue;
            }

            if (len(result.ClearTextPasswords) != 0) {
                test.Errorf("Case %d (%+v): User additions were disabled, but %d new cleartext passwords were found.", i, testCase, len(result.ClearTextPasswords));
                continue;
            }

            if (len(testMessages) != 0) {
                test.Errorf("Case %d (%+v): User additions were disabled, but %d new emails were found.", i, testCase, len(testMessages));
                continue;
            }
        }

        if (testCase.syncDel) {
            if (!usr.UsersPointerEqual(delUsers, result.Del)) {
                test.Errorf("Case %d (%+v): Unexpected del users. Expected: '%s', actual: '%s'.",
                        i, testCase, util.MustToJSON(delUsers), util.MustToJSON(result.Del));
                continue;
            }
        } else {
            if (len(result.Del) != 0) {
                test.Errorf("Case %d (%+v): User deletions were disabled, but %d deleted users were found.", i, testCase, len(result.Del));
                continue;
            }
        }
    }
}

// Get all possible test cases.
func getSyncLMSTestCases() []SyncLMSTestCase {
    return buildSyncLMSTestCase(nil, 0, make([]bool, 5));
}

func buildSyncLMSTestCase(testCases []SyncLMSTestCase, index int, currentCase []bool) []SyncLMSTestCase {
    if (index >= len(currentCase)) {
        return append(testCases, SyncLMSTestCase{
            dryRun: currentCase[0],
            sendEmails: currentCase[1],
            syncAttributes: currentCase[2],
            syncAdd: currentCase[3],
            syncDel: currentCase[4],
        });
    }

    currentCase[index] = true;
    testCases = buildSyncLMSTestCase(testCases, index + 1, currentCase);

    currentCase[index] = false;
    testCases = buildSyncLMSTestCase(testCases, index + 1, currentCase);

    return testCases;
}

// Reset the test LMS adapter back to it's starting settings.
func resetAdapter(course model.Course) {
    course.GetLMSAdapter().SyncUserAttributes = false;
    course.GetLMSAdapter().SyncAddUsers = false;
    course.GetLMSAdapter().SyncRemoveUsers = false;

    testLMSAdapter := course.GetLMSAdapter().Adapter.(*lmstest.TestLMSAdapter);
    testLMSAdapter.ClearUsersModifier();
}

// Modify the users that the LMS will return for testing.
func testingUsers(users []*lms.User) []*lms.User {
    // Remove other.
    removeIndex := -1;
    for i, user := range users {
        if (user.Email == "other@test.com") {
            removeIndex = i;
        } else if (user.Email == "student@test.com") {
            // student will only have their LMS ID added, no other changes.
        } else if (user.Email == "grader@test.com") {
            // grader will have their name changes.
            user.Name = "Changed Name";
        } else if (user.Email == "admin@test.com") {
            // admin will have their role changed.
            user.Role = usr.Owner;
        } else if (user.Email == "owner@test.com") {
            // owner will not have anything changed (so we must manually remove their LMS ID).
            user.ID = "";
        }
    }

    users = slices.Delete(users, removeIndex, removeIndex + 1);

    // Make an add user.
    addUser := &lms.User{
        ID: "lms-add@test.com",
        Name: "add",
        Email: "add@test.com",
        Role: usr.Student,
    };
    users = append(users, addUser);

    return users;
}

// The users that are marked as additions.
var addUsers []*usr.User = []*usr.User{
    &usr.User{
        Email: "add@test.com",
        DisplayName: "add",
        Role: usr.Student,
        LMSID: "lms-add@test.com",
    },
};

// The users that are marked as deletions.
var delUsers []*usr.User = []*usr.User{
    &usr.User{
        Email: "other@test.com",
        DisplayName: "other",
        Role: usr.Other,
        LMSID: "",
    },
};

// All the users that are marked as mods.
var modAllUsers []*usr.User = []*usr.User{
    &usr.User{
        Email: "student@test.com",
        DisplayName: "student",
        Role: usr.Student,
        LMSID: "lms-student@test.com",
    },
    &usr.User{
        Email: "grader@test.com",
        DisplayName: "Changed Name",
        Role: usr.Grader,
        LMSID: "lms-grader@test.com",
    },
    &usr.User{
        Email: "admin@test.com",
        DisplayName: "admin",
        Role: usr.Owner,
        LMSID: "lms-admin@test.com",
    },
};

// All the users that are marked as mods with no attribute syncing.
var modUsers []*usr.User = []*usr.User{
    &usr.User{
        Email: "student@test.com",
        DisplayName: "student",
        Role: usr.Student,
        LMSID: "lms-student@test.com",
    },
    &usr.User{
        Email: "grader@test.com",
        DisplayName: "grader",
        Role: usr.Grader,
        LMSID: "lms-grader@test.com",
    },
    &usr.User{
        Email: "admin@test.com",
        DisplayName: "admin",
        Role: usr.Admin,
        LMSID: "lms-admin@test.com",
    },
};

var addEmails []*email.Message = []*email.Message{
    &email.Message{
        To: []string{"add@test.com"},
        Subject: "Autograder -- User Account Created",
        HTML: false,
    },
};