package auerr

import (
	"fmt"

	"github.com/pkg/errors"
)

// ErrorInternalError unknown error for assetuploader app.
const ErrorInternalError = "ErrorInternalError"

// ErrorNotFound entity not found.
const ErrorNotFound = "ErrorNotFound"

// ErrorConflict operation on entity conflics with actual state
const ErrorConflict = "ErrorConflict"

// ErrorBadInput bad user input, validation error
const ErrorBadInput = "ErrorBadInput"

// SError creates a new error with a stacktrace and a msg.
func SError(code string, msg string) error {
	return errors.Wrap(errors.New(code), msg)
}

// FError creates a new error with a stacktrace and a formatted msg.
func FError(code string, msg string, vars ...interface{}) error {
	return errors.Wrap(errors.New(code), fmt.Sprintf(msg, vars...))
}

// CError creates a new error from another error, it also adds a stacktrace.
func CError(code string, err error) error {
	return errors.Wrap(errors.New(code), err.Error())
}
