package errs

import (
	"errors"
)

var (
	ErrFileExtensionNotAllowed error = errors.New("File extension is not allowed")
	ErrNoFileExtension         error = errors.New("File has no extension")

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
)
