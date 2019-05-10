package auerrors

import (
	"github.com/pkg/errors"
)

func New(code string, msg string) error {
	return errors.Wrap(errors.New(code), msg)
}

func NewWithError(code string, err error) error {
	return errors.Wrap(errors.New(code), err.Error())
}

