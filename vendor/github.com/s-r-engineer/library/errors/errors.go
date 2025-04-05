package libraryErrors

import (
	"errors"
	"fmt"

	libraryLogging "github.com/s-r-engineer/library/logging"
)

func Panicer(err any) {
	checker(err, func(a any) { panic(a) })
}

func Errorer(err error) {
	checker(err, func(a any) { libraryLogging.Error(a.(error).Error()) })
}

func checker(a any, f func(any)) {
	if a != nil {
		f(a)
	}
}

func PartWrapError(msg string) func(err error) error {
	return func(err error) error {
		return WrapError(msg, err)
	}
}

func PartWrapErrorOrString(msg string) (func(err error) error, func(err string) error) {
	return func(err error) error {
			return WrapError(msg, err)
		},
		func(err string) error {
			return WrapError(msg, errors.New(err))
		}
}

func WrapError(msg string, err error) error {
	if err != nil {
		return fmt.Errorf("%s -> %w", msg, err)
	}
	return nil
}

type LibError struct {
	msg string
}

func (e LibError) Error() string {
	return e.msg
}

func NewError(msg string) error {
	return LibError{msg: msg}
}

func NewErrorf(msg string, args ...any) error {
	return LibError{msg: fmt.Sprintf(msg, args...)}
}
