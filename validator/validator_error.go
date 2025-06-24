package validator

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ParseValidationErrors converts validator.ValidationErrors to []ValidationError
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
