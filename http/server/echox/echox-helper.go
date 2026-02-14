// Package echox provides helper utilities for the Echo web framework.
// It includes functions for request binding, validation, file handling,
// error responses, and query parameter parsing with pagination support.
package echox

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/apperror"
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/utils/primitive"
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/validatorx"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// defaultEchoxHelper is the default instance of EchoxHelper used by the Helper() function.
// It uses "message" as the JSON message key, "error_validations" as the validation error key,
// and 25 as the maximum page size.
var defaultEchoxHelper = NewEchoxHelper("message", "error_validations", 25, false)

// Helper returns the default EchoxHelper instance.
// This provides a convenient way to use EchoxHelper without creating a new instance.
func Helper() *EchoxHelper {
	return defaultEchoxHelper
}

// errKeyValue is the context key used to store error stack traces in the Echo context.
const errKeyValue string = "error_stack_trace_echox"

// EchoxHelper provides utility methods for handling common Echo web framework operations.
// It includes functionality for request binding, validation, file handling,
// error responses, and query parameter parsing.
type EchoxHelper struct {
	// keyJsonMessage is the JSON key used for error messages in responses
	keyJsonMessage string
	// keyErrorValidation is the JSON key used for validation errors in responses
	keyErrorValidation string
	// MaxPageSize is the maximum number of items per page for pagination
	MaxPageSize int64

	DebugMode bool
}

// NewEchoxHelper creates a new instance of EchoxHelper with custom configuration.
//
// Parameters:
//   - jsonMessage: the JSON key to use for error messages in responses
//   - errorValidation: the JSON key to use for validation errors in responses
//   - maxPageSize: the maximum number of items per page for pagination (defaults to 25 if 0)
//
// Returns:
//   - *EchoxHelper: a configured instance of EchoxHelper
//
// Example:
//
//	helper := NewEchoxHelper("msg", "errors", 50)
func NewEchoxHelper(jsonMessage string, errorValidation string, maxPageSize int64, debugMode bool) *EchoxHelper {
	if maxPageSize == 0 {
		maxPageSize = 25 // default value
	}
	return &EchoxHelper{
		keyJsonMessage:     jsonMessage,
		keyErrorValidation: errorValidation,
		MaxPageSize:        maxPageSize,
		DebugMode:          debugMode,
	}
}

// MustShouldBind attempts to bind and validate the request payload to the given struct.
// It handles both binding and validation in a single call, automatically writing
// appropriate error responses if either operation fails.
//
// The method performs the following steps:
//  1. Binds the request payload (JSON, form, query params, etc.) to the provided struct
//  2. Validates the bound struct using validator.v10 tags
//  3. Returns structured validation errors or binding errors as appropriate
//
// Parameters:
//   - c: Echo context containing the request
//   - req: pointer to the struct to bind the request data into
//
// Returns:
//   - bool: true if binding and validation succeeded, false if an error response was written
//
// HTTP Response Codes:
//   - 400 Bad Request: when validation fails (with detailed field errors) or binding fails
//   - 422 Unprocessable Entity: when non-validation errors occur
//
// Example:
//
//	type LoginRequest struct {
//	    Email    string `json:"email" validate:"required,email"`
//	    Password string `json:"password" validate:"required,min=8"`
//	}
//
//	func LoginHandler(c echo.Context) error {
//	    var req LoginRequest
//	    if !Helper().MustShouldBind(&c, &req) {
//	        return nil // error response already written
//	    }
//	    // proceed with valid request...
//	}
func (h *EchoxHelper) MustShouldBind(c *echo.Context, req any) bool {
	err := c.Bind(req)
	if err != nil {
		c.Set(errKeyValue, err.Error())

		_ = c.JSON(400, map[string]string{
			"message": err.Error(),
		})
	}

	if err := validatorx.Validate.StructCtx(c.Request().Context(), req); err != nil {
		c.Set(errKeyValue, err.Error())

		var verr validator.ValidationErrors
		if errors.As(err, &verr) {

			_ = c.JSON(http.StatusBadRequest, map[string]any{
				h.keyJsonMessage:     "Validation error",
				h.keyErrorValidation: validatorx.ParseValidationErrors(verr, "id"),
			})
			return false
		}

		_ = c.JSON(http.StatusUnprocessableEntity, map[string]any{
			h.keyJsonMessage: err.Error(),
		})
		return false
	}

	return true
}

// GetFiles retrieves multiple files from multipart form data.
// It extracts all files associated with the specified form field key and returns
// them as a slice of FileUpload structs, each containing the file name, size,
// and an open reader.
//
// Parameters:
//   - c: Echo context containing the multipart form request
//   - key: the form field name containing the files
//
// Returns:
//   - []primitive.FileUpload: slice of uploaded files with metadata
//   - error: error if form parsing fails or file cannot be opened
//
// Note: The caller is responsible for closing the File readers in each FileUpload.
//
// Example:
//
//	files, err := Helper().GetFiles(&c, "attachments")
//	if err != nil {
//	    return err
//	}
//	defer func() {
//	    for _, f := range files {
//	        f.File.Close()
//	    }
//	}()
func (h *EchoxHelper) GetFiles(c *echo.Context, key string) ([]primitive.FileUpload, error) {
	form, err := c.MultipartForm()
	if err != nil {
		c.Set(errKeyValue, err.Error())
		return nil, err
	}

	files := form.File[key]

	output := make([]primitive.FileUpload, 0, len(files))
	for _, file := range files {
		src, err := file.Open()
		if err != nil {
			c.Set(errKeyValue, err.Error())
			return nil, err
		}

		output = append(output, primitive.FileUpload{
			FileName: file.Filename,
			FileSize: file.Size,
			File:     src,
		})

	}

	return output, nil
}

// GetFile retrieves a single file from multipart form data.
// It extracts the file associated with the specified form field key and returns
// it as a FileUpload struct containing the file name, size, and an open reader.
//
// Parameters:
//   - c: Echo context containing the multipart form request
//   - key: the form field name containing the file
//
// Returns:
//   - primitive.FileUpload: the uploaded file with metadata
//   - error: error if form field not found, form parsing fails, or file cannot be opened
//
// Note: The caller is responsible for closing the File reader.
//
// Example:
//
//	file, err := Helper().GetFile(&c, "avatar")
//	if err != nil {
//	    return err
//	}
//	defer file.File.Close()
func (h *EchoxHelper) GetFile(c *echo.Context, key string) (primitive.FileUpload, error) {
	file, err := c.FormFile(key)
	if err != nil {
		c.Set(errKeyValue, err.Error())
		return primitive.FileUpload{}, err
	}

	src, err := file.Open()
	if err != nil {
		c.Set(errKeyValue, err.Error())
		return primitive.FileUpload{}, err
	}

	output := primitive.FileUpload{
		FileName: file.Filename,
		FileSize: file.Size,
		File:     src,
	}

	return output, nil
}

// ErrorResponse writes an appropriate HTTP error response based on the error type.
// It intelligently handles different error types and converts them to proper HTTP responses
// with appropriate status codes and messages.
//
// Behavior:
//   - If err is nil, returns nil without writing a response
//   - If err is *apperror.Error, extracts the HTTP code and public message
//   - For other errors, returns 500 Internal Server Error with a generic message
//   - Stores the full error details in the Echo context for logging/debugging
//
// Parameters:
//   - c: Echo context to write the response to
//   - err: the error to convert into an HTTP response
//
// Returns:
//   - error: the result of c.JSON() call, or nil if err was nil
//
// HTTP Response Format:
//
//	{
//	    "message": "error description"
//	}
//
// Example:
//
//	func Handler(c echo.Context) error {
//	    result, err := service.DoSomething()
//	    if err != nil {
//	        return Helper().ErrorResponse(&c, err)
//	    }
//	    return c.JSON(200, result)
//	}
func (h *EchoxHelper) ErrorResponse(c *echo.Context, err error) error {
	if err == nil {
		return nil
	}

	apperr, ok := apperror.As(err)
	httpCode := http.StatusInternalServerError
	msg := "Internal server error"
	stack := ""
	if ok {
		httpCode = apperr.Code.ToHTTPCode()
		msg = apperr.PublicMessage
		stack = apperr.Stack
	}
	c.Set(errKeyValue, err.Error())

	if h.DebugMode {
		return c.JSON(httpCode, map[string]string{
			h.keyJsonMessage: msg,
			"stack":          stack,
		})
	} else {
		return c.JSON(httpCode, map[string]string{
			h.keyJsonMessage: msg,
		})
	}
}

// QueryParamsToRangeDatePtr parses a date range query parameter into time pointers.
// The query parameter should contain one or two dates separated by a comma.
// If only one date is provided, only the start time is returned.
//
// Parameters:
//   - key: the query parameter name
//   - layout: the time format layout string (e.g., "2006-01-02" or time.RFC3339)
//   - c: Echo context containing the query parameters
//
// Returns:
//   - start: pointer to the start time, or nil if query param is empty
//   - end: pointer to the end time, or nil if not provided or query param is empty
//   - err: error if date parsing fails
//
// Query Parameter Format:
//   - Single date: "?date=2024-01-01" (only start is set)
//   - Date range: "?date=2024-01-01,2024-12-31" (both start and end are set)
//
// Example:
//
//	start, end, err := Helper().QueryParamsToRangeDatePtr("date_range", "2006-01-02", &c)
//	if err != nil {
//	    return err
//	}
//	if start != nil {
//	    // filter by start date
//	}
//	if end != nil {
//	    // filter by end date
//	}
func (h *EchoxHelper) QueryParamsToRangeDatePtr(key string, layout string, c *echo.Context) (start, end *time.Time, err error) {
	q := c.QueryParam(key)
	if q == "" {
		return
	}

	dateParts := strings.Split(q, ",")
	if len(dateParts) == 0 {
		return
	}

	startStr := dateParts[0]
	endStr := ""
	if len(dateParts) > 1 {
		endStr = dateParts[1]
	}

	startTime, err := time.Parse(layout, startStr)
	if err != nil {
		return start, end, errors.New("invalid start time query params")
	}

	if endStr != "" {
		endTime, err := time.Parse(layout, endStr)
		if err != nil {
			return start, end, errors.New("invalid end time query params")
		}

		end = &endTime
	}

	start = &startTime
	return start, end, nil
}

// QueryParamsToBooleanPtr parses a boolean query parameter into a bool pointer.
// This is useful for distinguishing between "false" and "not provided" in optional filters.
//
// Parameters:
//   - key: the query parameter name
//   - c: Echo context containing the query parameters
//
// Returns:
//   - value: pointer to the parsed boolean value, or nil if query param is empty
//   - err: error if the value cannot be parsed as a boolean
//
// Accepted Boolean Values:
//   - true: "1", "t", "T", "true", "TRUE", "True"
//   - false: "0", "f", "F", "false", "FALSE", "False"
//
// Example:
//
//	isActive, err := Helper().QueryParamsToBooleanPtr("is_active", &c)
//	if err != nil {
//	    return err
//	}
//	if isActive != nil {
//	    // filter by active status: *isActive is true or false
//	} else {
//	    // parameter not provided, don't filter by status
//	}
func (h *EchoxHelper) QueryParamsToBooleanPtr(key string, c *echo.Context) (value *bool, err error) {
	q := c.QueryParam(key)
	if q == "" {
		return
	}

	boolValue, err := strconv.ParseBool(q)
	if err != nil {
		c.Set(errKeyValue, err.Error())
		return nil, errors.New("invalid boolean query params")
	}

	value = &boolValue
	return value, nil
}

// QueryParamsToBoolean parses a boolean query parameter into a bool value.
// Unlike QueryParamsToBooleanPtr, this returns false (not nil) when the parameter is not provided.
//
// Parameters:
//   - key: the query parameter name
//   - c: Echo context containing the query parameters
//
// Returns:
//   - value: the parsed boolean value, or false if query param is empty
//   - err: error if the value cannot be parsed as a boolean
//
// Accepted Boolean Values:
//   - true: "1", "t", "T", "true", "TRUE", "True"
//   - false: "0", "f", "F", "false", "FALSE", "False"
//
// Example
//
//	includeDeleted, err := Helper().QueryParamsToBoolean("include_deleted", &c)
//	if err != nil {
//	    return err
//	}
//	// includeDeleted is false if not provided
func (h *EchoxHelper) QueryParamsToBoolean(key string, c *echo.Context) (value bool, err error) {
	q := c.QueryParam(key)
	if q == "" {
		return false, nil
	}

	value, err = strconv.ParseBool(q)
	if err != nil {
		c.Set(errKeyValue, err.Error())
		return false, errors.New("invalid boolean query params")
	}

	return value, nil
}

// BindToPaginationInput extracts pagination parameters from query parameters.
// It parses "page" and "page_size" query parameters and returns a PaginationInput struct
// with sensible defaults if the parameters are not provided or invalid.
//
// Parameters:
//   - c: Echo context containing the query parameters
//
// Returns:
//   - primitive.PaginationInput: struct containing page number and page size
//
// Default Values:
//   - page: 1 (if not provided or invalid)
//   - page_size: MaxPageSize from EchoxHelper config (default 25 if not provided or invalid)
//
// Query Parameters:
//   - page: the page number (starts from 1)
//   - page_size: the number of items per page
//
// Example:
//
//	func ListUsers(c echo.Context) error {
//	    pagination := Helper().BindToPaginationInput(&c)
//	    // pagination.Page will be 1 by default
//	    // pagination.PageSize will be 25 by default (or MaxPageSize configured)
//	    users, total := service.ListUsers(pagination)
//	    return c.JSON(200, map[string]any{
//	        "data":  users,
//	        "page":  pagination.Page,
//	        "total": total,
//	    })
//	}
func (h *EchoxHelper) BindToPaginationInput(c *echo.Context) primitive.PaginationInput {
	pagination := primitive.PaginationInput{
		Page:     1,
		PageSize: h.MaxPageSize,
	}

	page, _ := strconv.ParseInt(c.QueryParam("page"), 10, 64)
	if page != 0 {
		pagination.Page = page
	}
	pageSize, _ := strconv.ParseInt(c.QueryParam("page_size"), 10, 64)
	if pageSize != 0 {
		pagination.PageSize = pageSize
	}

	return pagination
}
