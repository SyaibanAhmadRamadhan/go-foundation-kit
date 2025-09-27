package ginx

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type NoopValidator struct{}

func (NoopValidator) ValidateStruct(any) error {
	return nil
}
func (NoopValidator) Engine() any {
	return nil
}

// GinConfig contains configuration for setting up a new Gin engine with
// OpenTelemetry middleware, custom validator, logging, CORS, etc.
type GinConfig struct {
	BlacklistRouteLogResponse map[string]struct{} // Routes that should not log response body
	SensitiveFields           map[string]struct{} // Fields that should be redacted from logs
	CorsConf                  CorsConfig          // CORS configuration
	AppName                   string              // Application name for OpenTelemetry tracing
	UseOtel                   bool
}

// NewGin creates and returns a configured *gin.Engine instance.
// It sets up recovery, CORS, OpenTelemetry tracing, logging, and validation.
func NewGin(conf GinConfig) *gin.Engine {
	if conf.SensitiveFields == nil {
		conf.SensitiveFields = make(map[string]struct{})
	}
	if conf.BlacklistRouteLogResponse == nil {
		conf.BlacklistRouteLogResponse = make(map[string]struct{})
	}

	router := gin.Default()

	ginValidator := &NoopValidator{}
	binding.Validator = ginValidator

	router.Use(gin.Recovery())
	router.Use(cors(conf.CorsConf))
	if conf.UseOtel {
		router.Use(otelgin.Middleware(conf.AppName))
	}
	router.Use(log(conf.BlacklistRouteLogResponse, conf.SensitiveFields))

	return router
}
