package apperror

import (
	"errors"
	"fmt"
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

	Cause error // <- tambah
}

func (e *Error) Unwrap() error { return e.Cause }

func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

func (e *Error) Error() string {
	if len(e.Stack) == 0 {
		return e.Message
	}

	msg := e.Message + "\n\nStack trace:\n" + e.Stack

	return msg
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

func StdUnknown(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, ErrUnknown) {
		return New(CodeUnknown, err.Error(),
			WithCause(err),
			WithPublicMessage("internal server error"),
		)
	}

	return New(CodeUnknown, err.Error(),
		WithCause(err),
		WithStack(),
		WithPublicMessage("internal server error"),
	)
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

func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }

func IsBadRequest(err error) bool { return errors.Is(err, ErrBadRequest) }

func IsUnauthorized(err error) bool { return errors.Is(err, ErrUnauthorized) }

func IsForbidden(err error) bool { return errors.Is(err, ErrForbidden) }

func IsConflict(err error) bool { return errors.Is(err, ErrConflict) }

func IsUnknown(err error) bool { return errors.Is(err, ErrUnknown) }

func As(err error) (*Error, bool) {
	var svcErr *Error
	ok := errors.As(err, &svcErr)
	return svcErr, ok
}

func CodeOf(err error) Code {
	e, ok := As(err)
	if ok {
		return e.Code
	}
	return CodeUnknown
}

func Format(err error, msg string) error {
	return fmt.Errorf("%s -> %w", msg, err)
}
