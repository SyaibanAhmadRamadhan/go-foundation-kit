package kafkax

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

// ErrProcessShutdownIsRunning is returned when a process shutdown is already in progress.
var ErrProcessShutdownIsRunning = errors.New("process shutdown is running")

// ErrJsonUnmarshal is returned when JSON unmarshalling fails.
var ErrJsonUnmarshal = errors.New("json unmarshal error")

// broker manages Kafka readers and writers, along with optional tracing integrations.
type broker struct {
	kafkaWriter   *kafka.Writer
	pubTracer     KafkaTracerPub
	consumeTracer KafkaTracerConsume
	commitTracer  KafkaTracerCommitMessage
	readers       []*Reader
}

// New creates a new broker instance with optional configurations applied via functional options.
// It also attempts to ping the Kafka writer during initialization.
//
// Parameters:
//   - opts: variadic list of Options to configure the broker
//
// Returns:
//   - *broker: initialized broker instance
func New(opts ...Options) *broker {
	b := &broker{
		readers: make([]*Reader, 0),
	}
	for _, option := range opts {
		option(b)
	}

	ctxPing, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	PingKafkaWriter(ctxPing, b.kafkaWriter)
	return b
}

// Close closes the Kafka writer and all registered readers.
// Logs any errors encountered during the closing process.
func (b *broker) Close() {
	var closeErrs []error

	if b.kafkaWriter != nil {
		if err := b.kafkaWriter.Close(); err != nil {
			slog.Error("failed to close kafka writer", slog.Any("error", err))
			closeErrs = append(closeErrs, fmt.Errorf("writer: %w", err))
		} else {
			slog.Info("kafka writer closed successfully")
		}
	}

	for i, reader := range b.readers {
		if reader == nil {
			continue
		}
		if err := reader.Close(); err != nil {
			slog.Error("failed to close kafka reader", slog.Int("index", i), slog.Any("error", err))
			closeErrs = append(closeErrs, fmt.Errorf("reader[%d]: %w", i, err))
		} else {
			slog.Info("kafka reader closed successfully", slog.Int("index", i))
		}
	}

	if len(closeErrs) > 0 {
		slog.Warn("broker close completed with errors", slog.Int("error_count", len(closeErrs)))
		for _, err := range closeErrs {
			slog.Warn("close error", slog.Any("error", err))
		}
	} else {
		slog.Info("broker closed cleanly without errors")
	}
}

// PingKafkaBrokers attempts to establish a TCP connection to each Kafka broker address
// using the provided Dialer to ensure connectivity.
//
// Parameters:
//   - ctx: context for timeout/cancellation
//   - brokers: list of broker addresses (host:port)
//   - dialer: kafka.Dialer instance to use for connection attempts
//
// Returns:
//   - error if any broker cannot be reached
func PingKafkaBrokers(ctx context.Context, brokers []string, dialer *kafka.Dialer) error {
	if len(brokers) == 0 {
		return fmt.Errorf("no Kafka brokers provided")
	}

	for _, addr := range brokers {
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to connect to Kafka broker %s: %w", addr, err)
		}
		_ = conn.Close()
	}
	return nil
}

// PingKafkaWriter attempts to dial the Kafka writer's broker address to verify availability.
//
// Parameters:
//   - ctx: context for timeout/cancellation
//   - writer: the kafka.Writer to be tested
//
// Returns:
//   - error if the writer is not reachable or dial fails
func PingKafkaWriter(ctx context.Context, writer *kafka.Writer) error {
	if writer == nil {
		return nil
	}

	addr := writer.Addr.String()

	conn, err := kafka.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to Kafka writer broker %s: %w", addr, err)
	}
	_ = conn.Close()
	return nil
}
