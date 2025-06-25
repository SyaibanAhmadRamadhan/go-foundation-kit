package observability

import (
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type OptionParams struct {
	ServiceName  string
	Env          string
	OtlpEndpoint string
	OtlpUsername string
	OtlpPassword string
}

func NewObservabilityOtel(params OptionParams) (trace.Tracer, func(), error) {
	closeFunc, err := NewOtel(params.ServiceName, params.OtlpEndpoint, params.OtlpUsername, params.OtlpPassword)
	if err != nil {
		return nil, nil, err
	}

	return otel.Tracer(params.ServiceName), func() {
		closeFunc()
	}, err
}

type LogWithKafkaHookOptions struct {
	KafkaAddrs  []string
	Transport   *kafka.Transport
	Topic       string
	Env         string
	ServiceName string
	LogMode     string // LogMode  "text" or "json"
	LogLevel    string // "info", "debug", etc.
	OnlySink    bool   // The log should only be sent to the sink (e.g., Kafka, file, etc.) and not printed to the terminal via slog, even if the environment is not "production".
}

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
