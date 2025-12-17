# AppError Package

A structured error handling package for Go applications that provides consistent error management with proper HTTP/gRPC code mappings and stack traces.

## Features

- 🎯 **Structured Error Types**: Well-defined error codes that map to HTTP and gRPC status codes
- 🔒 **Public/Internal Messages**: Separate messages for internal logging and public API responses
- 📍 **Stack Traces**: Optional stack trace collection for debugging
- 🌐 **Multi-Protocol Support**: Built-in HTTP and gRPC code mappings
- 🛡️ **Type Safety**: Strong typing with helper functions for common error scenarios
- 🔍 **Error Inspection**: Utilities to check error types and extract information

## Installation

```bash
go get github.com/SyaibanAhmadRamadhan/go-foundation-kit/apperror
```

## Error Codes

The package defines several application-level error codes:

| Code | HTTP Status | gRPC Code | Description |
|------|-------------|-----------|-------------|
| `CodeBadRequest` | 400 Bad Request | InvalidArgument | Invalid input data |
| `CodeUnauthorized` | 401 Unauthorized | Unauthenticated | Authentication required |
| `CodeForbidden` | 403 Forbidden | PermissionDenied | Access denied |
| `CodeNotFound` | 404 Not Found | NotFound | Resource not found |
| `CodeConflict` | 409 Conflict | Aborted | Resource conflict |
| `CodeUnknown` | 500 Internal Server Error | Internal | Unknown/Internal error |

## Quick Start

### Creating Errors

```go
package main

import (
    "fmt"
    "github.com/SyaibanAhmadRamadhan/go-foundation-kit/apperror"
)

func main() {
    // Simple error creation
    err := apperror.NotFound("user with ID 123 not found")

    // Error with custom public message
    err = apperror.BadRequest("invalid email format",
        apperror.WithPublicMessage("Please provide a valid email address"))

    // Error with stack trace
    err = apperror.Unknown("database connection failed",
        apperror.WithStack(),
        apperror.WithPublicMessage("Internal server error"))
}
```

### Helper Functions

```go
// Create specific error types
err := apperror.NotFound("resource not found")
err = apperror.BadRequest("invalid input")
err = apperror.Unauthorized("token expired")
err = apperror.Forbidden("access denied")
err = apperror.Conflict("resource already exists")

// Convert standard errors to structured errors
stdErr := errors.New("database error")
err = apperror.StdUnknown(stdErr) // Converts to CodeUnknown with stack trace
```

### Error Inspection

```go
func handleError(err error) {
    // Check if it's our structured error
    if apperror.Is(err) {
        appErr, _ := apperror.As(err)
        fmt.Printf("Code: %v, Message: %s, Public: %s\n",
            appErr.Code, appErr.Message, appErr.PublicMessage)
    }

    // Check specific error types
    if apperror.IsNotFound(err) {
        // Handle not found case
    }

    if apperror.IsBadRequest(err) {
        // Handle validation errors
    }

    // Get error code
    code := apperror.CodeOf(err)
    httpStatus := code.ToHTTPCode()
    grpcCode := code.ToGRPCCode()
}
```

## HTTP Integration

```go
func userHandler(w http.ResponseWriter, r *http.Request) {
    user, err := getUserByID("123")
    if err != nil {
        if appErr, ok := apperror.As(err); ok {
            http.Error(w, appErr.PublicMessage, appErr.Code.ToHTTPCode())
            return
        }
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Success case
    json.NewEncoder(w).Encode(user)
}
```

## Gin Integration Example

```go
func ginErrorHandler(c *gin.Context, err error) {
    if appErr, ok := apperror.As(err); ok {
        c.JSON(appErr.Code.ToHTTPCode(), gin.H{
            "message": appErr.PublicMessage,
            "error": appErr.Code,
        })
        return
    }

    c.JSON(http.StatusInternalServerError, gin.H{
        "message": "Internal server error",
    })
}
```

## gRPC Integration

```go
import "google.golang.org/grpc/status"

func grpcErrorHandler(err error) error {
    if appErr, ok := apperror.As(err); ok {
        return status.Error(appErr.Code.ToGRPCCode(), appErr.PublicMessage)
    }
    return status.Error(codes.Internal, "Internal server error")
}
```

## Configuration Options

### WithPublicMessage

Set a user-friendly message that's safe to expose in API responses:

```go
err := apperror.BadRequest("validation failed: email is required",
    apperror.WithPublicMessage("Email field is required"))
```

### WithStack

Include stack trace for debugging (useful for unknown/internal errors):

```go
err := apperror.Unknown("unexpected database error",
    apperror.WithStack())
```

### EnableStack

Conditionally enable stack traces:

```go
err := apperror.Unknown("error occurred",
    apperror.EnableStack(isDebugMode))
```

## Predefined Errors

The package includes common predefined errors:

```go
// Generic errors
apperror.ErrDataNotFound        // "data not found"
apperror.ErrDataAlreadyExists   // "data already exists"
apperror.ErrTransactionNotFound // "db transaction is null"
```

## Best Practices

1. **Use Specific Error Types**: Always use the most specific error code for your use case
2. **Separate Internal/Public Messages**: Keep internal messages detailed for logging, public messages user-friendly
3. **Include Stack Traces for Unknown Errors**: Help with debugging internal server errors
4. **Consistent Error Handling**: Use the helper functions for common scenarios
5. **Error Wrapping**: Convert standard errors using `StdUnknown()` when needed

## API Reference

### Constructor Functions

- `New(code Code, internalMsg string, opts ...Option) error`
- `NotFound(msg string, opts ...Option) error`
- `BadRequest(msg string, opts ...Option) error`
- `Unauthorized(msg string, opts ...Option) error`
- `Forbidden(msg string, opts ...Option) error`
- `Conflict(msg string, opts ...Option) error`
- `Unknown(msg string, opts ...Option) error`
- `StdUnknown(err error) error`

### Inspection Functions

- `Is(err error) bool` - Check if error is an apperror.Error
- `As(err error) (*Error, bool)` - Extract apperror.Error from error
- `CodeOf(err error) Code` - Get error code
- `IsNotFound(err error) bool`
- `IsBadRequest(err error) bool`
- `IsUnauthorized(err error) bool`
- `IsForbidden(err error) bool`
- `IsConflict(err error) bool`
- `IsUnknown(err error) bool`

### Options

- `WithPublicMessage(msg string) Option`
- `WithStack() Option`
- `EnableStack(enable bool) Option`

## License

This package is part of the go-foundation-kit project.
