package observability

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

// LogConfig contains configuration options for setting up structured logging.
type LogConfig struct {
	ZerologHook   io.Writer // Optional log sink (e.g., Kafka writer). If nil, logs won't be sent to external sinks.
	SlogHook      io.Writer // Optional log sink (e.g., Kafka writer). If nil, logs won't be sent to external sinks.
	Mode          string    // Output format: "text" or "json" for slog.
	Level         string    // Log level: "info", "debug", "warn", "error", etc.
	Env           string    // Environment name: "production", "staging", "development", etc.
	ServiceName   string    // The name of the service emitting the logs.
	ZerologStdOut bool
}

// NewLog initializes slog and zerolog based on the provided configuration.
// - If cfg.Hook is nil, logs will only be printed to stdout.
// - slog is used for local logs; zerolog is used for structured logs (e.g., Kafka).
func NewLog(cfg LogConfig) {
	buildSlog(cfg)
	buildZerolog(cfg)
}

func buildZerolog(cfg LogConfig) {
	hooks := make([]io.Writer, 0)
	if cfg.ZerologStdOut {
		hooks = append(hooks, os.Stdout)
	}
	if cfg.ZerologHook != nil {
		hooks = append(hooks, cfg.ZerologHook)
	}

	if len(hooks) > 0 {
		mw := zerolog.MultiLevelWriter(hooks...)
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		zerologLevel := parseZerologLevel(cfg.Level)

		zlogger := zerolog.New(mw).Level(zerologLevel).With().
			Timestamp().
			Str("env", cfg.Env).
			Str("service_name", cfg.ServiceName).
			Logger()

		log.Logger = zlogger
	}
}

// buildSlog returns a slog.Handler configured with the specified mode and log level.
func buildSlog(cfg LogConfig) {
	slogLevel := parseSlogLevel(cfg.Level)

	if cfg.SlogHook == nil {
		cfg.SlogHook = os.Stdout
	}
	opts := &slog.HandlerOptions{Level: slogLevel}
	var slogHandler slog.Handler
	switch strings.ToLower(cfg.Mode) {
	case "json":
		slogHandler = slog.NewJSONHandler(cfg.SlogHook, opts)
	default:
		slogHandler = slog.NewTextHandler(cfg.SlogHook, opts)
	}

	slog.SetDefault(slog.New(slogHandler))
	slog.Info("slog initialized",
		slog.String("env", cfg.Env),
		slog.String("service", cfg.ServiceName),
		slog.String("mode", cfg.Mode),
		slog.String("level", cfg.Level),
	)
}

// parseSlogLevel converts a string log level into a slog.Level.
// Defaults to slog.LevelInfo if level is unrecognized.
func parseSlogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// parseZerologLevel converts a string log level into a zerolog.Level.
// Defaults to zerolog.InfoLevel if level is unrecognized.
func parseZerologLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// Start returns a zerolog event enriched with OpenTelemetry trace and span IDs extracted from context.
// This is useful for logging within distributed tracing environments.
//
// Example:
//
//	log := observability.Start(ctx, zerolog.InfoLevel)
//	log.Msg("some message")
func Start(ctx context.Context, level zerolog.Level) *zerolog.Event {
	traceID := ""
	spanID := ""
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		traceID = spanContext.TraceID().String()
		spanID = spanContext.SpanID().String()
	}

	switch level {
	case zerolog.TraceLevel:
		return log.Trace().Str("trace_id", traceID).Str("span_id", spanID)
	case zerolog.DebugLevel:
		return log.Debug().Str("trace_id", traceID).Str("span_id", spanID)
	case zerolog.InfoLevel:
		return log.Info().Str("trace_id", traceID).Str("span_id", spanID)
	case zerolog.WarnLevel:
		return log.Warn().Str("trace_id", traceID).Str("span_id", spanID)
	case zerolog.ErrorLevel:
		return log.Error().Str("trace_id", traceID).Str("span_id", spanID)
	case zerolog.FatalLevel:
		return log.Fatal().Str("trace_id", traceID).Str("span_id", spanID)
	case zerolog.PanicLevel:
		return log.Panic().Str("trace_id", traceID).Str("span_id", spanID)
	default:
		return log.Info().Str("trace_id", traceID).Str("span_id", spanID)
	}
}
