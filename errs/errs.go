package errs

import (
	"errors"
)

var (
	ErrFileExtensionNotAllowed error = errors.New("File extension is not allowed")
	ErrNoFileExtension         error = errors.New("File has no extension")
	ErrEmptyFileName           error = errors.New("Filename can't be empty")
	ErrEmptyName               error = errors.New("Name can't be empty")

	ErrNotAdmin           error = errors.New("Admin permission required")
	ErrNotModerator       error = errors.New("Moderator permissions required")
	ErrNotUser            error = errors.New("User permissions required")
	ErrNotCourseAdmin     error = errors.New("Course admin permissions required")
	ErrNotCourseModerator error = errors.New("Course moderator permissions required")
	ErrNotCourseUser      error = errors.New("Course user permissions required")

	ErrParameterConversion error = errors.New("Unable to convert parameter item")
	ErrNoQuery             error = errors.New("Unable to find query parameter")
	ErrRawData             error = errors.New("Unable to get raw data from request")
	ErrNoFileInRequest     error = errors.New("Unable to find file in request")
	ErrBodyConversion      error = errors.New("Unable to convert body")

	ErrSelfRegisterExam         error = errors.New("Cannot register for own exam")
	ErrRegisterDeadlinePassed   error = errors.New("Cannot register to exam past deadline")
	ErrUnregisterDeadlinePassed error = errors.New("Cannot unregister from exam past deadline")
	ErrExamHasntStarted         error = errors.New("Exam hasn't started yet")
	ErrExamEnded                error = errors.New("Exam already ended")

	ErrNoUploads          error = errors.New("This item doesn't have any associated uplods")
	ErrUploadLimitReached error = errors.New("The upload limit has been reached")

	ErrCourseNotEmpty error = errors.New("Course is not empty")
	ErrWrongEnrollkey error = errors.New("Wrong enroll key")

	ErrVisibleTimePast             error = errors.New("VisibleFrom time can't be in the past")
	ErrDeadlineTimePast            error = errors.New("Deadline time can't be in the past")
	ErrVisibleFromAfterDeadline    error = errors.New("Visible from time can't be after deadline")
	ErrSubmissionTimeAfterDeadline error = errors.New("Submission time is past deadline time of this submission")
)
