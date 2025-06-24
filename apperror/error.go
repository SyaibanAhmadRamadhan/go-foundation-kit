package apperror

import (
	"errors"
	"net/http"
)

var ErrDataNotFound = errors.New("data not found")
var ErrDataAlreadyExists = errors.New("data already exists")
var ErrTransactionNotFound = errors.New("db transaction is null")

type Error struct {
	Message string
	Code    int
}

func (e *Error) Error() string {
	return e.Message
}

// Factory
func NewError(code int, msg string) error {
	return &Error{
		Message: msg,
		Code:    code,
	}
}

// Shorthand
func ErrNotFound(msg string) error {
	return NewError(http.StatusNotFound, msg)
}

func ErrConflict(msg string) error {
	return NewError(http.StatusConflict, msg)
}

func ErrBadRequest(msg string) error {
	return NewError(http.StatusBadRequest, msg)
}

func ErrUnauthorized(msg string) error {
	return NewError(http.StatusUnauthorized, msg)
}

func ErrForbidden(msg string) error {
	return NewError(http.StatusForbidden, msg)
}

func ErrorsAsNotFound(err error) bool {
	var svcErr *Error
	if errors.As(err, &svcErr) {
		return svcErr.Code == http.StatusNotFound
	}
	return false
}
