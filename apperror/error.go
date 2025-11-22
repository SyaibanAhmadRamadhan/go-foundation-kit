package apperror

import (
	"errors"
)

var (
	// ErrDataNotFound is a generic error indicating that requested data was not found.
	ErrDataNotFound = errors.New("data not found")

	// ErrDataAlreadyExists indicates that the data already exists in the system.
	ErrDataAlreadyExists = errors.New("data already exists")

	// ErrTransactionNotFound indicates that a database transaction was expected but not found.
	ErrTransactionNotFound = errors.New("db transaction is null")
)

// Error represents a structured application error with a generic code,
// internal message, public message, and optional stack trace.
type Error struct {
	// internal message (buat logging)
	Message string

	// aman ditampilkan ke user
	PublicMessage string

	Code  Code
	Stack string
}

func (e *Error) Error() string {
	if e.Stack != "" {
		return e.Message + " | stack: " + e.Stack
	}
	return e.Message
}

func New(code Code, internalMsg string, opts ...Option) error {
	e := &Error{
		Message: internalMsg,
		Code:    code,
	}

	for _, opt := range opts {
		opt(e)
	}

	if e.PublicMessage == "" {
		e.PublicMessage = e.Message
	}

	return e
}

// Helpers dipakai di service layer
func NotFound(msg string, opts ...Option) error {
	return New(CodeNotFound, msg, opts...)
}

func Unknown(msg string, opts ...Option) error {
	return New(CodeUnknown, msg, opts...)
}

func StdUnknown(err Error) error {
	return New(CodeUnknown, err.Error(), WithStack(), WithPublicMessage("internal server error"))
}

func Conflict(msg string, opts ...Option) error {
	return New(CodeConflict, msg, opts...)
}

func BadRequest(msg string, opts ...Option) error {
	return New(CodeBadRequest, msg, opts...)
}

func Unauthorized(msg string, opts ...Option) error {
	return New(CodeUnauthorized, msg, opts...)
}

func Forbidden(msg string, opts ...Option) error {
	return New(CodeForbidden, msg, opts...)
}

// Predicates
func hasCode(err error, code Code) bool {
	var svcErr *Error
	return errors.As(err, &svcErr) && svcErr.Code == code
}

func IsNotFound(err error) bool {
	return hasCode(err, CodeNotFound)
}
func IsBadRequest(err error) bool {
	return hasCode(err, CodeBadRequest)
}
func IsUnauthorized(err error) bool {
	return hasCode(err, CodeUnauthorized)
}
func IsForbidden(err error) bool {
	return hasCode(err, CodeForbidden)
}
func IsConflict(err error) bool {
	return hasCode(err, CodeConflict)
}
func IsUnknown(err error) bool {
	var svcErr *Error
	if !errors.As(err, &svcErr) {
		return true
	}
	return svcErr.Code == CodeUnknown
}
func As(err error) (*Error, bool) {
	var svcErr *Error
	ok := errors.As(err, &svcErr)
	return svcErr, ok
}

func Is(err error) bool {
	_, ok := As(err)
	return ok
}

func CodeOf(err error) Code {
	e, ok := As(err)
	if ok {
		return e.Code
	}
	return CodeUnknown
}
