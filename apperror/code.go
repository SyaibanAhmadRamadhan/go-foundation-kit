package apperror

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

type Code int

const (
	CodeUnknown Code = iota
	CodeBadRequest
	CodeUnauthorized
	CodeForbidden
	CodeNotFound
	CodeConflict
)

// ToHTTPCode converts application error code to HTTP status code.
func (c Code) ToHTTPCode() int {
	switch c {
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// ToGRPCCode converts application error code to gRPC standard codes.
func (c Code) ToGRPCCode() codes.Code {
	switch c {
	case CodeBadRequest:
		return codes.InvalidArgument
	case CodeUnauthorized:
		return codes.Unauthenticated
	case CodeForbidden:
		return codes.PermissionDenied
	case CodeNotFound:
		return codes.NotFound
	case CodeConflict:
		return codes.Aborted
	default:
		return codes.Internal
	}
}

var (
	ErrNotFound     = &Error{Code: CodeNotFound}
	ErrBadRequest   = &Error{Code: CodeBadRequest}
	ErrUnauthorized = &Error{Code: CodeUnauthorized}
	ErrForbidden    = &Error{Code: CodeForbidden}
	ErrConflict     = &Error{Code: CodeConflict}
	ErrUnknown      = &Error{Code: CodeUnknown}
)
