package ginx

import (
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/utils/primitive"
	"github.com/gin-gonic/gin"
)

// Default GinHelper instance with standard JSON keys
var defaultHelper = NewGinHelper("message", "error_validations")

// Convenience functions that use the default helper instance.
// These provide backward compatibility and ease of use for common cases.

// MustShouldBind attempts to bind the request payload to the given struct.
// Uses default JSON keys: "message" and "error_validations".
// - If validation fails (validator.v10), it returns 400 with structured errors.
// - If binding/parsing fails (JSON/form decode, type mismatch, etc.), it returns 422 with the raw error.
// Returns false if response is already written; true if everything is OK.
func MustShouldBind(c *gin.Context, req any) bool {
	return defaultHelper.MustShouldBind(c, req)
}

// ErrorResponse writes an error response to the context using default JSON keys.
// If the error is of type *apperror.Error, it uses the associated HTTP code.
// Internal server errors are masked with a generic message.
func ErrorResponse(c *gin.Context, err error) {
	defaultHelper.ErrorResponse(c, err)
}

// MustParseQueryToSliceInt64 parses a comma-separated string query value into a slice of int64.
// If the value is empty or nil, it returns nil.
// If parsing fails, it returns 0 for invalid values.
func MustParseQueryToSliceInt64(value *string) []int64 {
	return defaultHelper.MustParseQueryToSliceInt64(value)
}

// MustParseQueryToSliceFloat64 parses a comma-separated string query value into a slice of float64.
// If the value is empty or nil, it returns nil.
// If parsing fails, it returns 0 for invalid values.
func MustParseQueryToSliceFloat64(value *string) []float64 {
	return defaultHelper.MustParseQueryToSliceFloat64(value)
}

// ParseQueryToSliceString parses a comma-separated string query value into a slice of strings.
// If the value is empty or nil, it returns nil.
func ParseQueryToSliceString(value *string) []string {
	return defaultHelper.ParseQueryToSliceString(value)
}

// BindToPaginationInput extracts pagination parameters from the context.
// If parameters are not set, it defaults to page=1 and pageSize=25.
func BindToPaginationInput(c *gin.Context) primitive.PaginationInput {
	return defaultHelper.BindToPaginationInput(c)
}
