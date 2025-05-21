package db

import (
	"fmt"
	"slices"
	"strings"

	"github.com/edulinq/autograder/internal/model"
)

// Resolve course email addresses.
// Take a course and a list of strings (containing emails specs) and returns a sorted slice of lowercase emails without duplicates.
// An email spec can be:
// an email address,
// a course role (which will include all course users with that role),
// a literal "*" (which includes all users enrolled in the course),
// or an email address preceded by a dash ("-") (which indicates that this email address should NOT be included in the final results).
func ResolveCourseUsers(course *model.Course, emails []model.CourseUserReferenceInput) ([]string, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	reference, err := ParseCourseUserReference(course, emails)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse input references: '%w'.", err)
	}

	emailSet := reference.Emails

	// Add or remove users based off their course role.
	if len(reference.CourseUserRoles) > 0 || len(reference.ExcludeCourseUserRoles) > 0 {
		users, err := GetCourseUsers(course)
		if err != nil {
			return nil, err
		}

		for _, user := range users {
			// Remove a user if their role is set to exclude.
			_, ok := reference.ExcludeCourseUserRoles[user.Role.String()]
			if ok {
				delete(emailSet, strings.ToLower(user.Email))
				continue
			}

			// Add a user if their role is set.
			_, ok = reference.CourseUserRoles[user.Role.String()]
			if ok {
				emailSet[strings.ToLower(user.Email)] = nil
			}
		}
	}

	// Remove negative users.
	for email, _ := range reference.ExcludeEmails {
		delete(emailSet, email)
	}

	emailSlice := make([]string, 0, len(emailSet))
	for email := range emailSet {
		emailSlice = append(emailSlice, email)
	}

	slices.Sort(emailSlice)

	return emailSlice, nil
}
