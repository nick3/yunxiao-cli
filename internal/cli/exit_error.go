package cli

import "errors"

type ExitError struct {
	Code    int
	Message string
}

func NewExitError(code int, message string) *ExitError {
	return &ExitError{Code: code, Message: message}
}

func (e *ExitError) Error() string {
	return e.Message
}

func ExitCode(err error) (int, bool) {
	var exitErr *ExitError
	if errors.As(err, &exitErr) {
		return exitErr.Code, true
	}
	return ExitSuccess, false
}
