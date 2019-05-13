package auerr

import (
	"fmt"

	"github.com/pkg/errors"
)

const ErrorInternalError = "ErrorInternalError"
const ErrorNotFound = "ErrorNotFound"
const ErrorConflict = "ErrorConflict"

func SError(code string, msg string) error {
	return errors.Wrap(errors.New(code), msg)
}
func FError(code string, msg string, vars ...interface{}) error {
	return errors.Wrap(errors.New(code), fmt.Sprintf(msg, vars))
}
func CError(code string, err error) error {
	return errors.Wrap(errors.New(code), err.Error())
}
