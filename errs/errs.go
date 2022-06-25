package errs

import "errors"

var (
	ErrFileExtensionNotAllowed error = errors.New("File extension is not allowed")
	ErrNoFileExtension         error = errors.New("File has no extension")
)
