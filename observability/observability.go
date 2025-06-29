package observability

import (
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// OptionParams defines the configuration for setting up OpenTelemetry observability.
type OptionParams struct {
	ServiceName  string // The name of the service used for tracing.
	Env          string // The environment (e.g., "development", "staging", "production").
	OtlpEndpoint string // The OTLP exporter endpoint.
	OtlpUsername string // The username for OTLP authentication.
	OtlpPassword string // The password for OTLP authentication.
}

// NewObservabilityOtel initializes OpenTelemetry tracing and returns a tracer and cleanup function.
//
// It returns:
//   - trace.Tracer: the tracer instance to be used for creating spans.
//   - func(): a cleanup function that should be called before shutdown.
//   - error: if initialization fails.
func NewObservabilityOtel(params OptionParams) (trace.Tracer, func(), error) {
	closeFunc, err := NewOtel(params.ServiceName, params.OtlpEndpoint, params.OtlpUsername, params.OtlpPassword)
	if err != nil {
		return nil, nil, err
	}

	return otel.Tracer(params.ServiceName), func() {
		closeFunc()
	}, err
}

// LogWithKafkaHookOptions contains configuration for setting up logging with a Kafka sink.
type LogWithKafkaHookOptions struct {
	KafkaAddrs  []string         // List of Kafka broker addresses.
	Transport   *kafka.Transport // Optional custom transport (e.g., with TLS, SASL).
	Topic       string           // Kafka topic to write logs to.
	Env         string           // The environment (e.g., "production", "development").
	ServiceName string           // The name of the service emitting logs.
	LogMode     string           // Log output format: "text" or "json".
	LogLevel    string           // Minimum log level: "info", "debug", etc.
	OnlySink    bool             // If true, logs are only sent to the sink and not printed to terminal.
}

// NewLogWithKafkaHook sets up structured logging using slog with a Kafka writer as the hook.
// It returns a cleanup function that should be deferred or called on shutdown to close the Kafka writer.
func NewLogWithKafkaHook(optionsParams LogWithKafkaHookOptions) func() {
	w := &kafka.Writer{
		Addr:            kafka.TCP(optionsParams.KafkaAddrs...),
		Topic:           optionsParams.Topic,
		Balancer:        &kafka.LeastBytes{},
		MaxAttempts:     5,
		WriteBackoffMin: time.Duration(100),
		WriteBackoffMax: time.Duration(1 * time.Second),

		BatchSize:    10,
		BatchBytes:   1048576,
		BatchTimeout: time.Duration(3 * time.Second),

		RequiredAcks: kafka.RequireOne,
		Transport:    optionsParams.Transport,
	}
	NewLog(LogConfig{
		Hook: &KafkaHook{
			writer:      w,
			topic:       optionsParams.Topic,
			env:         optionsParams.Env,
			serviceName: optionsParams.ServiceName,
			onlySink:    optionsParams.OnlySink,
		},
		Mode:        optionsParams.LogMode,
		Level:       optionsParams.LogLevel,
		Env:         optionsParams.Env,
		ServiceName: optionsParams.ServiceName,
	})

	return func() {
		slog.Info("shutting down kafka writer...",
			slog.String("topic", optionsParams.Topic),
			slog.String("env", optionsParams.Env),
			slog.String("service", optionsParams.ServiceName),
		)

		if err := w.Close(); err != nil {
			slog.Error("failed to close kafka writer",
				slog.String("topic", optionsParams.Topic),
				slog.Any("error", err),
			)
		} else {
			slog.Info("kafka writer closed successfully",
				slog.String("topic", optionsParams.Topic),
			)
		}
	}
}
