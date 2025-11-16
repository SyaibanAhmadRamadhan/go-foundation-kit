package chix

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/observability"
	"github.com/rs/zerolog"
)

type stackTraceKey uint

const (
	stackTraceKeyCtx stackTraceKey = 0
)

type bodyWriter struct {
	http.ResponseWriter

	status      int
	logRespBody bool
	body        *bytes.Buffer
}

func (w bodyWriter) Write(b []byte) (int, error) {
	if w.logRespBody {
		w.body.Write(b)
	}

	return w.ResponseWriter.Write(b)
}

func (sr *bodyWriter) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
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

func log(blacklistRouteLogResponse map[string]struct{}, sensitiveFields map[string]struct{}) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			method := r.Method
			path := r.URL.Path
			key := method + " " + path

			query := r.URL.Query()
			reqQueryParams := make(map[string]any, len(query))
			for key, values := range query {
				if len(values) == 1 {
					reqQueryParams[key] = values[0]
				} else {
					reqQueryParams[key] = values
				}
			}

			reqBody := make(map[string]any)
			if r.Body != nil && r.ContentLength > 0 {
				bodyBytes, err := io.ReadAll(r.Body)
				if err == nil {
					if json.Unmarshal(bodyBytes, &reqBody) == nil {
						redactSensitiveFields(reqBody, sensitiveFields)
					}
					r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}

			var respBody map[string]any
			_, ok := blacklistRouteLogResponse[key]
			blw := &bodyWriter{
				body:           bytes.NewBuffer([]byte{}),
				ResponseWriter: w,
				logRespBody:    !ok,
			}
			w = blw
			next.ServeHTTP(w, r)

			if !ok {
				contentType := w.Header().Get("Content-Type")
				if strings.Contains(contentType, "application/json") {
					if json.Unmarshal(blw.body.Bytes(), &respBody) == nil {
						redactSensitiveFields(respBody, sensitiveFields)
					}
				}
			}

			status := blw.status
			level := zerolog.InfoLevel
			switch {
			case status >= 500:
				level = zerolog.ErrorLevel
			case status >= 400:
				level = zerolog.WarnLevel
			}

			e := observability.Start(r.Context(), level).
				Str("method", method).
				Str("path", path).
				Int("status_code", status)

			err, ok := r.Context().Value(stackTraceKeyCtx).(string)
			if ok && len(err) > 0 {
				e.Str("error", err)
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
		})
	}
}
