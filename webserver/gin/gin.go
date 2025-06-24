package libgin

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type ginValidatorCustom struct {
	validator *validator.Validate
}

// ValidateStruct is called by Gin to validate the struct
func (cv *ginValidatorCustom) ValidateStruct(obj any) error {
	if err := cv.validator.Struct(obj); err != nil {
		return err
	}
	return nil
}

// Engine is called by Gin to retrieve the underlying validation engine
func (cv *ginValidatorCustom) Engine() any {
	return cv.validator
}

type GinConfig struct {
	BlacklistRouteLogResponse map[string]struct{}
	SensitiveFields           map[string]struct{}
	Validator                 *validator.Validate
	CorsConf                  CorsConfig
	AppName                   string
}

func NewGin(conf GinConfig) *gin.Engine {
	router := gin.Default()

	ginValidator := &ginValidatorCustom{
		validator: conf.Validator,
	}
	binding.Validator = ginValidator
	router.Use(gin.Recovery())
	router.Use(cors(conf.CorsConf))
	router.Use(otelgin.Middleware(conf.AppName))
	router.Use(trace(conf.BlacklistRouteLogResponse, conf.SensitiveFields))
	return router
}
