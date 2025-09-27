package chix

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Config contains configuration for setting up a new Gin engine with
// OpenTelemetry middleware, custom validator, logging, CORS, etc.
type Config struct {
	BlacklistRouteLogResponse map[string]struct{} // Routes that should not log response body
	SensitiveFields           map[string]struct{} // Fields that should be redacted from logs
	CorsConf                  CorsConfig          // CORS configuration
	AppName                   string              // Application name for OpenTelemetry tracing
	UseOtel                   bool
}

func New(conf Config) *chi.Mux {
	if conf.SensitiveFields == nil {
		conf.SensitiveFields = make(map[string]struct{})
	}
	if conf.BlacklistRouteLogResponse == nil {
		conf.BlacklistRouteLogResponse = make(map[string]struct{})
	}
	r := chi.NewRouter()
	r.Use(cors(conf.CorsConf))
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)

	r.Use(log(conf.BlacklistRouteLogResponse, conf.SensitiveFields))
	if conf.UseOtel {
		r.Use(func(next http.Handler) http.Handler {
			return otelhttp.NewHandler(next, conf.AppName)
		})
	}

	return r
}
