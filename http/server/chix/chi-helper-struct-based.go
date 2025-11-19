package chix

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/apperror"
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/utils/primitive"
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/validatorx"
	"github.com/go-playground/validator/v10"
)

// ChiHelper is a small utility wrapper used in HTTP handlers (typically with the chi router)
// to standardize JSON response keys such as "message" and "error_validations".
// This helps maintain consistent JSON output format across the application.
type ChiHelper struct {
	keyJsonMessage     string
	keyErrorValidation string
}

// NewChiHelper returns a new ChiHelper instance.
//
// If keyJsonMessage is empty, it defaults to "message".
// If keyErrorValidation is empty, it defaults to "error_validations".
//
// This allows the caller to override JSON response keys while still providing
// sensible defaults when not specified.
func NewChiHelper(keyJsonMessage string, keyErrorValidation string) *ChiHelper {
	if keyJsonMessage == "" {
		keyJsonMessage = "message"
	}
	if keyErrorValidation == "" {
		keyErrorValidation = "error_validations"
	}

	return &ChiHelper{
		keyJsonMessage:     keyJsonMessage,
		keyErrorValidation: keyErrorValidation,
	}
}

// MustShouldBindJSON attempts to bind JSON/form payload to src and validate it.
// - Returns 400 with structured errors if validation fails.
// - Returns 422 with raw error if decoding/binding fails.
// - Returns false if response already written; true if OK.
func (h *ChiHelper) MustShouldBindJSON(w http.ResponseWriter, r *http.Request, src any) (bool, *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		r = SetError(r, err)
		Write(w, http.StatusUnprocessableEntity, "application/json", map[string]any{
			h.keyJsonMessage: "failed to read body: " + err.Error(),
		})
		return false, r
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, src); err != nil {
		r = SetError(r, err)
		Write(w, http.StatusUnprocessableEntity, "application/json", map[string]any{
			h.keyJsonMessage: err.Error(),
		})
		return false, r
	}

	if err := validatorx.Validate.StructCtx(r.Context(), src); err != nil {
		r = SetError(r, err)

		var verr validator.ValidationErrors
		if errors.As(err, &verr) {
			Write(w, http.StatusBadRequest, "application/json", map[string]any{
				h.keyJsonMessage:     "Validation error",
				h.keyErrorValidation: validatorx.ParseValidationErrors(verr),
			})
			return false, r
		}

		// error lain saat validasi
		Write(w, http.StatusUnprocessableEntity, "application/json", map[string]any{
			h.keyJsonMessage: err.Error(),
		})
		return false, r
	}

	return true, r
}

func (h *ChiHelper) SetError(r *http.Request, err error) *http.Request {
	existingError := r.Context().Value(stackTraceKeyCtx).(string)
	if existingError != "" {
		existingError += " | " + err.Error()
	}

	return r.WithContext(context.WithValue(r.Context(), stackTraceKeyCtx, existingError))
}

// helper untuk menulis JSON response
func (h *ChiHelper) Write(w http.ResponseWriter, status int, contentType string, v any) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// ErrorResponse writes an error response to the context.
// If the error is of type *apperror.Error, it uses the associated HTTP code.
// Internal server errors are masked with a generic message.
func (h *ChiHelper) ErrorResponse(w http.ResponseWriter, r *http.Request, err error) *http.Request {
	if err == nil {
		return r
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

	r = SetError(r, err)
	Write(w, httpCode, "application/json", map[string]string{
		h.keyJsonMessage: msg,
	})

	return r
}

// ParseQueryToSliceInt64 parses a comma-separated string query value into a slice of int64.
// If the value is empty or nil, it returns nil.
func (h *ChiHelper) ParseQueryToSliceInt64(value *string) ([]int64, error) {
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
func (h *ChiHelper) ParseQueryToSliceFloat64(value *string) ([]float64, error) {
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
func (h *ChiHelper) ParseQueryToSliceString(value *string) ([]string, error) {
	if value == nil || *value == "" {
		return nil, nil
	}
	return strings.Split(*value, ","), nil
}

// BindToPaginationInput extracts pagination params from URL query.
// Defaults: page=1, page_size=25.
// Accepts both "page_size" and "pageSize".
func (h *ChiHelper) BindToPaginationInput(r *http.Request) primitive.PaginationInput {
	const (
		defaultPage     int64 = 1
		defaultPageSize int64 = 25
	)

	q := r.URL.Query()

	page := parseInt64OrDefault(q.Get("page"), defaultPage)
	if page < 1 {
		page = defaultPage
	}

	psRaw := q.Get("page_size")
	if psRaw == "" {
		psRaw = q.Get("pageSize")
	}
	pageSize := parseInt64OrDefault(psRaw, defaultPageSize)
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	return primitive.PaginationInput{
		Page:     page,
		PageSize: pageSize,
	}
}

func (h *ChiHelper) parseInt64OrDefault(s string, def int64) int64 {
	if s == "" {
		return def
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return n
}
