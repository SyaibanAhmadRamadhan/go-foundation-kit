package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

// KafkaHook is a custom log writer that sends log entries to a Kafka topic.
// It can optionally print logs to the terminal in non-production environments.
type KafkaHook struct {
	Writer      *kafka.Writer // Kafka writer instance.
	Topic       string        // Kafka topic where logs will be published.
	Env         string        // Current environment (e.g., "production", "staging", "development").
	ServiceName string        // Name of the service generating logs.
	OnlySink    bool          // If true, logs are only sent to the sink and not printed to the terminal.
}

// Write implements the io.Writer interface for KafkaHook.
// It sends the log `p` to Kafka, optionally printing it to the terminal based on environment.
//
// Log payload is expected to be in JSON format.
// It enriches Kafka headers with fields like level, trace_id, span_id, and status_code (if present).
func (w *KafkaHook) Write(p []byte) (n int, err error) {
	if w.Env != "production" && !w.OnlySink {
		slog.Info("log output (non-production)",
			slog.String("service", w.ServiceName),
			slog.String("env", w.Env),
			slog.String("log", string(p)),
		)
	}
	var payload map[string]any
	if err := json.Unmarshal(p, &payload); err != nil {
		slog.Error("KafkaLogWriter: failed to parse log JSON",
			slog.String("service", w.ServiceName),
			slog.String("env", w.Env),
			slog.Any("error", err),
			slog.String("raw", string(p)),
		)
		return len(p), nil
	}

	level := payload["level"]
	statusCode := payload["status_code"]
	spanID := payload["span_id"]
	traceID := payload["trace_id"]
	headers := []kafka.Header{
		{Key: "service_name", Value: []byte(w.ServiceName)},
		{Key: "env", Value: []byte(w.Env)},
		{Key: "level", Value: fmt.Appendf(nil, "%v", level)},
	}

	if statusCode != nil {
		headers = append(headers, kafka.Header{
			Key: "status_code", Value: fmt.Appendf(nil, "%v", statusCode),
		})
	}
	if spanID != nil {
		headers = append(headers, kafka.Header{
			Key: "span_id", Value: fmt.Appendf(nil, "%v", spanID),
		})
	}
	if traceID != nil {
		headers = append(headers, kafka.Header{
			Key: "trace_id", Value: fmt.Appendf(nil, "%v", traceID),
		})
	}

	if traceID == nil {
		slog.Warn("KafkaLogWriter: log entry missing trace_id",
			slog.String("service", w.ServiceName),
			slog.String("env", w.Env),
			slog.Any("payload", payload),
		)
		return
	}

	err = w.Writer.WriteMessages(context.Background(), kafka.Message{
		Value:   p,
		Headers: headers,
	})
	if err != nil {
		slog.Error("KafkaLogWriter: failed to send log to Kafka",
			slog.String("service", w.ServiceName),
			slog.String("env", w.Env),
			slog.Any("error", err),
			slog.String("trace_id", fmt.Sprintf("%v", traceID)),
		)
	}

	return len(p), nil
}
