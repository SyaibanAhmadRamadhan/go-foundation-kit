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
// It implements the KafkaBroker interface.
type broker struct {
	pubTracer     TracerPub
	consumeTracer TracerConsume
	commitTracer  TracerCommitMessage

	writers map[string]*kafka.Writer
	readers []*Reader
}

// Ensure broker implements KafkaPubSub interface
var _ PubSub = (*broker)(nil)

// New creates a new broker instance with optional configurations applied via functional options.
// It also attempts to ping the Kafka writer during initialization.
//
// Parameters:
//   - opts: variadic list of Options to configure the broker
//
// Returns:
//   - KafkaBroker: initialized broker instance
func New(opts ...Options) (*broker, error) {
	b := &broker{
		readers: make([]*Reader, 0),
	}
	for _, option := range opts {
		option(b)
	}

	ctxPing, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := PingWriters(ctxPing, b.writers)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Close closes the Kafka writer and all registered readers.
// Returns an error if any close operation fails.
func (b *broker) Close() error {
	var closeErrs []error

	if len(b.writers) <= 0 {
		for _, writer := range b.writers {
			if err := writer.Close(); err != nil {
				slog.Error("failed to close kafka writer", slog.Any("error", err))
				closeErrs = append(closeErrs, fmt.Errorf("writer: %w", err))
			} else {
				slog.Info("kafka writer closed successfully")
			}
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
		return fmt.Errorf("multiple close errors: %v", closeErrs)
	}

	slog.Info("broker closed cleanly without errors")
	return nil
}

func (b *broker) GetWriter(key string) (*kafka.Writer, error) {
	if w, ok := b.writers[key]; ok {
		return w, nil
	}

	return nil, errors.New("unknown writer")
}

// PingBrokers attempts to establish a TCP connection to each Kafka broker address
// using the provided Dialer to ensure connectivity.
//
// Parameters:
//   - ctx: context for timeout/cancellation
//   - brokers: list of broker addresses (host:port)
//   - dialer: kafka.Dialer instance to use for connection attempts
//
// Returns:
//   - error if any broker cannot be reached
func PingBrokers(ctx context.Context, brokers []string, dialer *kafka.Dialer) error {
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

// PingWriters attempts to dial the Kafka writer's broker address to verify availability.
//
// Parameters:
//   - ctx: context for timeout/cancellation
//   - writer: the kafka.Writer to be tested
//
// Returns:
//   - error if the writer is not reachable or dial fails
func PingWriters(ctx context.Context, writers map[string]*kafka.Writer) error {
	if len(writers) <= 0 {
		return nil
	}

	for _, writer := range writers {
		addr := writer.Addr.String()

		conn, err := kafka.DialContext(ctx, "tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to connect to Kafka writer broker %s: %w", addr, err)
		}
		_ = conn.Close()
	}
	return nil
}
