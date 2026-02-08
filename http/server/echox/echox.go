package echox

import (
	"github.com/labstack/echo/v5"
	emiddleware "github.com/labstack/echo/v5/middleware"
)

type Config struct {
	BlacklistRouteLogResponse map[string]struct{} // Routes that should not log response body
	SensitiveFields           map[string]struct{} // Fields that should be redacted from logs
	AppName                   string              // Application name for OpenTelemetry tracing
	CorsConfig                emiddleware.CORSConfig
}

func New(conf Config) *echo.Echo {
	e := echo.New()

	e.Use(emiddleware.Recover())
	e.Use(emiddleware.CORSWithConfig(conf.CorsConfig))
	e.Use(log(conf.BlacklistRouteLogResponse, conf.SensitiveFields))
	return e
}
