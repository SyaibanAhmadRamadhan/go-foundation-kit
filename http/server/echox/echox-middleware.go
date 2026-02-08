package echox

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/observability"
	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
)

type bodyLogWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func redactSensitiveFields(data map[string]any, sensitiveFields map[string]struct{}) {
	for key, val := range data {
		if _, ok := sensitiveFields[key]; ok {
			data[key] = "[REDACTED]"
			continue
		}

		switch typed := val.(type) {
		case map[string]any:
			redactSensitiveFields(typed, sensitiveFields)

		case []any:
			for _, item := range typed {
				if m, ok := item.(map[string]any); ok {
					redactSensitiveFields(m, sensitiveFields)
				}
			}
		}
	}
}

func log(blacklistRouteLogResponse map[string]struct{}, sensitiveFields map[string]struct{}) echo.MiddlewareFunc {
	const maxBodySize = 1 << 20 // 1MB

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			res := c.Response()

			method := req.Method
			path := c.Path()
			key := method + ":" + path

			// Query params
			queryParams := c.QueryParams()
			reqQueryParams := make(map[string]any, len(queryParams))
			for k, values := range queryParams {
				if len(values) == 1 {
					reqQueryParams[k] = values[0]
				} else {
					reqQueryParams[k] = values
				}
			}

			// Request body (JSON only)
			var reqBody map[string]any
			ctReq := req.Header.Get("Content-Type")
			if strings.HasPrefix(ctReq, "application/json") && req.Body != nil {
				limited := io.LimitReader(req.Body, maxBodySize+1)
				bodyBytes, err := io.ReadAll(limited)
				if err != nil {
					return err
				}
				if len(bodyBytes) > maxBodySize {
					return echo.NewHTTPError(http.StatusRequestEntityTooLarge, "request body too large")
				}

				req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

				if len(bodyBytes) > 0 {
					var tmp map[string]any
					if err := json.Unmarshal(bodyBytes, &tmp); err == nil {
						redactSensitiveFields(tmp, sensitiveFields)
						reqBody = tmp
					}
				}
			}

			var respBody map[string]any
			_, blacklisted := blacklistRouteLogResponse[key]

			blw := &bodyLogWriter{
				body:           bytes.NewBuffer(nil),
				ResponseWriter: res,
			}
			c.SetResponse(blw)

			err := next(c)

			// parse response body kalau perlu
			if !blacklisted {
				ctResp := res.Header().Get("Content-Type")
				if strings.Contains(ctResp, "application/json") {
					if json.Unmarshal(blw.body.Bytes(), &respBody) == nil {
						redactSensitiveFields(respBody, sensitiveFields)
					}
				}
			}

			status := blw.statusCode
			level := zerolog.InfoLevel
			switch {
			case status >= 500:
				level = zerolog.ErrorLevel
			case status >= 400:
				level = zerolog.WarnLevel
			}

			e := observability.Start(req.Context(), level).
				Str("method", method).
				Str("path", path).
				Int("status_code", status)

			if err != nil {
				e.Str("error", err.Error())
			}

			if respBody != nil {
				e.Any("response_body", respBody)
			}
			if reqBody != nil {
				e.Any("request_body", reqBody)
			}
			if len(reqQueryParams) > 0 {
				e.Any("query_parameters", reqQueryParams)
			}

			errVal := c.Get(errKeyValue)
			if errVal != nil {
				if errInCtx, ok := errVal.(string); ok {
					e.Str("error", errInCtx)
				}
			}

			e.Msg("HTTP Request")
			// status code yang benar

			return err
		}
	}
}
