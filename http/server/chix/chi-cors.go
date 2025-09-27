package chix

import (
	"net/http"
	"strings"
)

type CorsConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
}

func cors(config CorsConfig) func(next http.Handler) http.Handler {
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

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowOrigins)
			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
			w.Header().Set("Access-Control-Allow-Methods", allowMethods)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
