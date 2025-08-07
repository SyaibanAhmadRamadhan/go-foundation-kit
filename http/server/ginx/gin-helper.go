package ginx

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/apperror"
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/utils/primitive"
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/validator"
	"github.com/gin-gonic/gin"
)

// MustShouldBind attempts to bind the request payload to the given struct.
// If validation fails, it returns a 400 response with validation errors.
// Otherwise, it returns a 422 with the raw error.
func MustShouldBind(c *gin.Context, req any) bool {
	if err := c.ShouldBind(req); err != nil {
		c.Error(err)
		validationErr := validator.ParseValidationErrors(err)
		if len(validationErr) > 0 {
			c.JSON(http.StatusBadRequest, map[string]any{
				"message":           "Validation error",
				"error_validations": validationErr,
			})
			return false
		}
		c.JSON(http.StatusUnprocessableEntity, map[string]string{
			"message": err.Error(),
		})
		return false
	}
	return true
}

// ErrorResponse writes an error response to the context.
// If the error is of type *apperror.Error, it uses the associated HTTP code.
// Internal server errors are masked with a generic message.
func ErrorResponse(c *gin.Context, err error) {
	if err == nil {
		return
	}
	var apperr *apperror.Error
	httpCode := http.StatusInternalServerError
	msg := "Internal server error"
	if errors.As(err, &apperr) {
		httpCode = apperr.Code
		if httpCode >= http.StatusInternalServerError {
			msg = "Internal server error"
		} else {
			switch httpCode {
			case http.StatusUnauthorized:
				msg = "unauthorized"
			default:
				msg = apperr.Error()
			}
		}
	}
	c.Error(err)
	c.JSON(httpCode, map[string]string{
		"message": msg,
	})
}

// ParseQueryToSliceInt64 parses a comma-separated string query value into a slice of int64.
// If the value is empty or nil, it returns nil.
func ParseQueryToSliceInt64(value *string) ([]int64, error) {
	if value == nil || *value == "" {
		return nil, nil
	}

	values := strings.Split(*value, ",")
	intValues := make([]int64, len(values))
	for i, v := range values {
		intValue, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, apperror.ErrBadRequest("Invalid query parameter")
		}
		intValues[i] = intValue
	}
	return intValues, nil
}

// ParseQueryToSliceFloat64 parses a comma-separated string query value into a slice of float64.
// If the value is empty or nil, it returns nil.
func ParseQueryToSliceFloat64(value *string) ([]float64, error) {
	if value == nil || *value == "" {
		return nil, nil
	}
	values := strings.Split(*value, ",")
	floatValues := make([]float64, len(values))
	for i, v := range values {
		floatValue, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, apperror.ErrBadRequest("Invalid query parameter")
		}
		floatValues[i] = floatValue
	}
	return floatValues, nil
}

// ParseQueryToSliceString parses a comma-separated string query value into a slice of strings.
// If the value is empty or nil, it returns nil.
func ParseQueryToSliceString(value *string) ([]string, error) {
	if value == nil || *value == "" {
		return nil, nil
	}
	return strings.Split(*value, ","), nil
}

// BindToPaginationInput extracts pagination parameters from the context.
// If parameters are not set, it defaults to page=1 and pageSize=25.
func BindToPaginationInput(c *gin.Context) primitive.PaginationInput {
	pagination := primitive.PaginationInput{
		Page:     1,
		PageSize: 25,
	}

	page := c.GetInt64("page")
	if page != 0 {
		pagination.Page = page
	}
	pageSize := c.GetInt64("page_size")
	if pageSize != 0 {
		pagination.PageSize = pageSize
	}

	return pagination
}
