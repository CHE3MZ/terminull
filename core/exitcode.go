package core

import (
	"context"
	"errors"
)

// ExitCodeError carries a process exit code alongside the cause.
// Main functions map any returned error via ExitCodeOf and os.Exit.
type ExitCodeError struct {
	code int
	err  error
}

// NewExitCodeError builds an ExitCodeError with the given code and cause.
func NewExitCodeError(code int, err error) ExitCodeError {
	return ExitCodeError{
		code: code,
		err:  err,
	}
}

// ExitCode returns the process exit code.
func (e ExitCodeError) ExitCode() int {
	return e.code
}

// Error returns the cause message.
func (e ExitCodeError) Error() string {
	if e.err == nil {
		return ""
	}

	return e.err.Error()
}

// Unwrap returns the cause so errors.Is/As see through it.
func (e ExitCodeError) Unwrap() error {
	return e.err
}

// ExitCodeOf maps any error to a process exit code:
// the carried code for ExitCodeError (even when wrapped),
// 130 for a cancelled context (Ctrl+C), 1 for anything else, 0 for nil.
func ExitCodeOf(err error) int {
	if err == nil {
		return 0
	}

	var exitErr ExitCodeError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}

	if errors.Is(err, context.Canceled) {
		return 130
	}

	return 1
}
