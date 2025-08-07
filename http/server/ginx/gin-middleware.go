package ginx

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/observability"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func redactSensitiveFields(data map[string]any, sensitive map[string]struct{}) {
	for key, val := range data {
		if _, ok := sensitive[key]; ok {
			data[key] = "[REDACTED]"
			continue
		}

		switch typed := val.(type) {
		case map[string]any:
			redactSensitiveFields(typed, sensitive)

		case []any:
			for _, item := range typed {
				if m, ok := item.(map[string]any); ok {
					redactSensitiveFields(m, sensitive)
				}
			}
		}
	}
}

func trace(blacklistRouteLogResponse map[string]struct{}, sensitiveFields map[string]struct{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.FullPath()
		key := method + ":" + path

		query := c.Request.URL.Query()
		reqQueryParams := make(map[string]any, len(query))
		for key, values := range query {
			if len(values) == 1 {
				reqQueryParams[key] = values[0]
			} else {
				reqQueryParams[key] = values
			}
		}

		reqBody := make(map[string]any)
		if c.Request.Body != nil && c.Request.ContentLength > 0 {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				if json.Unmarshal(bodyBytes, &reqBody) == nil {
					redactSensitiveFields(reqBody, sensitiveFields)
				}
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		var respBody map[string]any
		_, ok := blacklistRouteLogResponse[key]
		if ok {
			c.Next()
		} else {
			blw := &bodyLogWriter{body: bytes.NewBuffer([]byte{}), ResponseWriter: c.Writer}
			c.Writer = blw

			c.Next()

			contentType := c.Writer.Header().Get("Content-Type")
			if strings.Contains(contentType, "application/json") {
				if json.Unmarshal(blw.body.Bytes(), &respBody) == nil {
					redactSensitiveFields(respBody, sensitiveFields)
				}
			}
		}

		status := c.Writer.Status()
		level := zerolog.InfoLevel
		switch {
		case status >= 500:
			level = zerolog.ErrorLevel
		case status >= 400:
			level = zerolog.WarnLevel
		}

		e := observability.Start(c.Request.Context(), level).
			Str("method", method).
			Str("path", path).
			Int("status_code", status)

		if len(c.Errors) > 0 {
			e.Str("error", c.Errors.String())
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

		e.Msg("HTTP Request")
	}
}

type CorsConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
}

func cors(config CorsConfig) gin.HandlerFunc {
	allowOrigins := "*"
	if len(config.AllowOrigins) > 0 {
		allowOrigins = strings.Join(config.AllowOrigins, ", ")
	}

	allowMethods := "POST, OPTIONS, GET, PUT, PATCH, DELETE"
	if len(config.AllowMethods) > 0 {
		allowMethods = strings.Join(config.AllowMethods, ", ")
	}

	allowHeaders := "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"
	if len(config.AllowHeaders) > 0 {
		allowHeaders = strings.Join(config.AllowHeaders, ", ")
	}

	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigins)
		if config.AllowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		c.Writer.Header().Set("Access-Control-Allow-Headers", allowHeaders)
		c.Writer.Header().Set("Access-Control-Allow-Methods", allowMethods)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
