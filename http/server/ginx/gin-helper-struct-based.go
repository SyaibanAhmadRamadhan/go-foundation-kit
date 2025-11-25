package ginx

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/apperror"
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/utils/primitive"
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/validatorx"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// GinHelper is a small utility wrapper used in HTTP handlers (typically with the gin router)
// to standardize JSON response keys such as "message" and "error_validations".
// This helps maintain consistent JSON output format across the application.
type GinHelper struct {
	keyJsonMessage     string
	keyErrorValidation string
}

// NewGinHelper returns a new GinHelper instance.
//
// If keyJsonMessage is empty, it defaults to "message".
// If keyErrorValidation is empty, it defaults to "error_validations".
//
// This allows the caller to override JSON response keys while still providing
// sensible defaults when not specified.
func NewGinHelper(keyJsonMessage string, keyErrorValidation string) *GinHelper {
	if keyJsonMessage == "" {
		keyJsonMessage = "message"
	}
	if keyErrorValidation == "" {
		keyErrorValidation = "error_validations"
	}

	return &GinHelper{
		keyJsonMessage:     keyJsonMessage,
		keyErrorValidation: keyErrorValidation,
	}
}

// MustShouldBind attempts to bind the request payload to the given struct.
// - If validation fails (validator.v10), it returns 400 with structured errors.
// - If binding/parsing fails (JSON/form decode, type mismatch, etc.), it returns 422 with the raw error.
// Returns false if response is already written; true if everything is OK.
func (h *GinHelper) MustShouldBind(c *gin.Context, req any) bool {
	if err := c.ShouldBind(req); err != nil {
		c.Error(err)

		var verr validator.ValidationErrors
		if errors.As(err, &verr) {
			c.JSON(http.StatusBadRequest, map[string]any{
				h.keyJsonMessage:     "Validation error",
				h.keyErrorValidation: validatorx.ParseValidationErrors(verr),
			})
			return false
		}

		c.JSON(http.StatusUnprocessableEntity, map[string]any{
			h.keyJsonMessage: err.Error(),
		})
		return false
	}

	if err := validatorx.Validate.StructCtx(c.Request.Context(), req); err != nil {
		c.Error(err)

		var verr validator.ValidationErrors
		if errors.As(err, &verr) {
			c.JSON(http.StatusBadRequest, map[string]any{
				h.keyJsonMessage:     "Validation error",
				h.keyErrorValidation: validatorx.ParseValidationErrors(verr),
			})
			return false
		}

		// error lain saat validasi (jarang, tapi jaga-jaga)
		c.JSON(http.StatusUnprocessableEntity, map[string]any{
			h.keyJsonMessage: err.Error(),
		})
		return false
	}

	return true
}

// ErrorResponse writes an error response to the context.
// If the error is of type *apperror.Error, it uses the associated HTTP code.
// Internal server errors are masked with a generic message.
func (h *GinHelper) ErrorResponse(c *gin.Context, err error) {
	if err == nil {
		return
	}
	apperr, ok := apperror.As(err)
	httpCode := http.StatusInternalServerError
	msg := "Internal server error"
	if ok {
		httpCode = apperr.Code.ToHTTPCode()
		if httpCode >= http.StatusInternalServerError {
			msg = "Internal server error"
		} else {
			msg = apperr.PublicMessage
		}
	}
	c.Error(err)
	c.JSON(httpCode, map[string]string{
		h.keyJsonMessage: msg,
	})
}

// ParseQueryToSliceInt64 parses a comma-separated string query value into a slice of int64.
// If the value is empty or nil, it returns nil.
func (h *GinHelper) ParseQueryToSliceInt64(value *string) ([]int64, error) {
	if value == nil || *value == "" {
		return nil, nil
	}

	values := strings.Split(*value, ",")
	intValues := make([]int64, len(values))
	for i, v := range values {
		intValue, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, apperror.BadRequest("Invalid query parameter")
		}
		intValues[i] = intValue
	}
	return intValues, nil
}

// ParseQueryToSliceFloat64 parses a comma-separated string query value into a slice of float64.
// If the value is empty or nil, it returns nil.
func (h *GinHelper) ParseQueryToSliceFloat64(value *string) ([]float64, error) {
	if value == nil || *value == "" {
		return nil, nil
	}
	values := strings.Split(*value, ",")
	floatValues := make([]float64, len(values))
	for i, v := range values {
		floatValue, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, apperror.BadRequest("Invalid query parameter")
		}
		floatValues[i] = floatValue
	}
	return floatValues, nil
}

// ParseQueryToSliceString parses a comma-separated string query value into a slice of strings.
// If the value is empty or nil, it returns nil.
func (h *GinHelper) ParseQueryToSliceString(value *string) ([]string, error) {
	if value == nil || *value == "" {
		return nil, nil
	}
	return strings.Split(*value, ","), nil
}

// BindToPaginationInput extracts pagination parameters from the context.
// If parameters are not set, it defaults to page=1 and pageSize=25.
func (h *GinHelper) BindToPaginationInput(c *gin.Context) primitive.PaginationInput {
	pagination := primitive.PaginationInput{
		Page:     1,
		PageSize: 25,
	}

	page, _ := strconv.ParseInt(c.Query("page"), 10, 64)
	if page != 0 {
		pagination.Page = page
	}
	pageSize, _ := strconv.ParseInt(c.Query("page_size"), 10, 64)
	if pageSize != 0 {
		pagination.PageSize = pageSize
	}

	return pagination
}
