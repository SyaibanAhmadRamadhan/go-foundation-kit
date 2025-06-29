package apperror

import (
	"errors"
	"net/http"
)

var (
	// ErrDataNotFound is a generic error indicating that requested data was not found.
	ErrDataNotFound = errors.New("data not found")

	// ErrDataAlreadyExists indicates that the data already exists in the system.
	ErrDataAlreadyExists = errors.New("data already exists")

	// ErrTransactionNotFound indicates that a database transaction was expected but not found.
	ErrTransactionNotFound = errors.New("db transaction is null")
)

// Error represents a structured application error with a message and HTTP status code.
type Error struct {
	Message string
	Code    int
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// NewError creates a new application-level error with the given HTTP status code and message.
func NewError(code int, msg string) error {
	return &Error{
		Message: msg,
		Code:    code,
	}
}

// ErrNotFound creates a new 404 Not Found error with the specified message.
func ErrNotFound(msg string) error {
	return NewError(http.StatusNotFound, msg)
}

// ErrConflict creates a new 409 Conflict error with the specified message.
func ErrConflict(msg string) error {
	return NewError(http.StatusConflict, msg)
}

// ErrBadRequest creates a new 400 Bad Request error with the specified message.
func ErrBadRequest(msg string) error {
	return NewError(http.StatusBadRequest, msg)
}

// ErrUnauthorized creates a new 401 Unauthorized error with the specified message.
func ErrUnauthorized(msg string) error {
	return NewError(http.StatusUnauthorized, msg)
}

// ErrForbidden creates a new 403 Forbidden error with the specified message.
func ErrForbidden(msg string) error {
	return NewError(http.StatusForbidden, msg)
}

// ErrorsAsNotFound returns true if the provided error is an application-level
// error and has a 404 Not Found HTTP status code.
func ErrorsAsNotFound(err error) bool {
	var svcErr *Error
	if errors.As(err, &svcErr) {
		return svcErr.Code == http.StatusNotFound
	}
	return false
}
