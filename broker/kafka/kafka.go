package libkafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/segmentio/kafka-go"
)

var ErrProcessShutdownIsRunning = errors.New("process shutdown is running")
var ErrJsonUnmarshal = errors.New("json unmarshal error")

type broker struct {
	kafkaWriter  *kafka.Writer
	pubTracer    TracerPub
	subTracer    TracerSub
	commitTracer TracerCommitMessage
	readers      []*Reader
}

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

func findOwnImportedVersion() {
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		for _, dep := range buildInfo.Deps {
			if dep.Path == TracerName {
				kafkaLibVersion = dep.Version
			}
		}
	}
}

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
