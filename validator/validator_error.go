package validator

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

// ValidationError represents a simplified error response
// for a specific field when validation fails.
//
// Fields:
//   - Field: the name of the struct field that failed validation
//   - Message: the translated validation error message for that field
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ParseValidationErrors converts a validator.ValidationErrors (from go-playground/validator)
// into a slice of custom ValidationError structs for easier consumption (e.g., in JSON responses).
//
// Parameters:
//   - err: the original error returned by validator.Validate.Struct(...)
//
// Returns:
//   - []ValidationError: a slice of structured validation errors, or nil if the input error is not a validation error.
//
// Notes:
//   - Each ValidationError includes the field name and a translated message.
//   - It uses the global `TranslatorID`, which must be previously initialized with a language translator.
func ParseValidationErrors(err error) []ValidationError {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		var result []ValidationError
		for _, e := range ve {
			result = append(result, ValidationError{
				Field:   e.Field(),
				Message: e.Translate(TranslatorID),
			})
		}
		return result
	}
	return nil
}
