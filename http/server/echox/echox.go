// Package echox provides a wrapper around the Echo web framework with additional features
// such as automatic healthcheck endpoint registration, logging, and CORS configuration.
package echox

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
	emiddleware "github.com/labstack/echo/v5/middleware"
)

// HealthCheckFunc is a function type for health check operations.
// It should return an error if the health check fails, or nil if healthy.
// The error message will be included in the health check response.
type HealthCheckFunc func(ctx context.Context) error

// HealthCheckItem represents a single health check with its name and check function.
type HealthCheckItem struct {
	// Name is the identifier for this health check (e.g., "database", "redis", "kafka")
	Name string
	// Check is the function that performs the actual health check
	Check HealthCheckFunc
}

// Config holds the configuration for creating a new Echox instance.
type Config struct {
	// BlacklistRouteLogResponse contains routes that should not log response body
	BlacklistRouteLogResponse map[string]struct{}
	// SensitiveFields contains field names that should be redacted from logs
	SensitiveFields map[string]struct{}
	// AppName is the application name for OpenTelemetry tracing
	AppName string
	// CorsConfig is the CORS configuration for the Echo server
	CorsConfig emiddleware.CORSConfig
	// HealthChecks contains the list of health check functions to register
	// If provided, a healthcheck endpoint will be automatically registered at /health
	HealthChecks []HealthCheckItem
	// HealthCheckPath is the custom path for the health check endpoint
	// Defaults to "/health" if not specified
	HealthCheckPath string
}

// Echox is a wrapper around Echo framework with additional features.
type Echox struct {
	e *echo.Echo
}

// New creates a new Echox instance with the provided configuration.
// It automatically sets up:
//   - Panic recovery middleware
//   - CORS middleware
//   - Request/response logging middleware
//   - Health check endpoint (if HealthChecks are provided)
//
// Parameters:
//   - conf: Configuration for the Echox instance
//
// Returns:
//   - *Echox: A configured Echox instance
//
// Example:
//
//	echox := New(Config{
//	    AppName: "my-service",
//	    CorsConfig: middleware.CORSConfig{
//	        AllowOrigins: []string{"*"},
//	    },
//	    HealthChecks: []HealthCheckItem{
//	        {
//	            Name: "database",
//	            Check: func(ctx context.Context) error {
//	                return db.PingContext(ctx)
//	            },
//	        },
//	        {
//	            Name: "redis",
//	            Check: func(ctx context.Context) error {
//	                return redisClient.Ping(ctx).Err()
//	            },
//	        },
//	    },
//	})
func New(conf Config) *Echox {
	e := echo.New()

	e.Use(log(conf.BlacklistRouteLogResponse, conf.SensitiveFields))
	e.Use(emiddleware.Recover())
	e.Use(emiddleware.CORSWithConfig(conf.CorsConfig))

	// Register health check endpoint if health checks are provided
	if len(conf.HealthChecks) > 0 {
		healthPath := conf.HealthCheckPath
		if healthPath == "" {
			healthPath = "/health"
		}
		e.GET(healthPath, createHealthCheckHandler(conf.HealthChecks))
	}

	return &Echox{
		e: e,
	}
}

// Echo returns the underlying Echo instance.
// This allows direct access to the Echo instance for advanced configuration.
//
// Returns:
//   - *echo.Echo: The underlying Echo instance
func (c *Echox) Echo() *echo.Echo {
	return c.e
}

// createHealthCheckHandler creates an HTTP handler that executes all registered health checks.
// It returns a JSON response with the overall status and individual check results.
//
// Response format:
//
//	{
//	    "status": "healthy" | "unhealthy",
//	    "checks": {
//	        "check_name": {
//	            "status": "pass" | "fail",
//	            "error": "error message if failed"
//	        }
//	    }
//	}
//
// HTTP Status Codes:
//   - 200 OK: All health checks passed
//   - 503 Service Unavailable: One or more health checks failed
func createHealthCheckHandler(checks []HealthCheckItem) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()

		type checkResult struct {
			Status string `json:"status"`
			Error  string `json:"error,omitempty"`
		}

		results := make(map[string]checkResult)
		allHealthy := true

		// Execute all health checks
		for _, item := range checks {
			err := item.Check(ctx)
			if err != nil {
				allHealthy = false
				results[item.Name] = checkResult{
					Status: "fail",
					Error:  err.Error(),
				}
			} else {
				results[item.Name] = checkResult{
					Status: "pass",
				}
			}
		}

		// Determine overall status
		overallStatus := "healthy"
		statusCode := http.StatusOK
		if !allHealthy {
			overallStatus = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		}

		return c.JSON(statusCode, map[string]interface{}{
			"status": overallStatus,
			"checks": results,
		})
	}
}
