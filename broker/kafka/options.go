package libkafka

import (
	"github.com/segmentio/kafka-go"
)

// Options defines a function that configures the broker during initialization.
// This is part of the functional options pattern commonly used in Go.
type Options func(cfg *broker)

// KafkaWriter creates an option to initialize a Kafka writer using the given
// broker addresses and topic.
//
// Parameters:
//   - url: list of Kafka broker addresses (e.g., []string{"localhost:9092"})
//   - topic: the Kafka topic to which messages will be written
//
// The writer will use the LeastBytes balancer strategy to distribute messages.
func KafkaWriter(url []string, topic string) Options {
	return func(cfg *broker) {
		cfg.kafkaWriter = &kafka.Writer{
			Addr:     kafka.TCP(url...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		}
	}
}

// KafkaCustomWriter allows you to provide a fully customized *kafka.Writer instance.
//
// Useful if you want fine-grained control over Kafka writer configuration such as
// retries, batch size, compression, async, etc.
func KafkaCustomWriter(k *kafka.Writer) Options {
	return func(cfg *broker) {
		cfg.kafkaWriter = k
	}
}

// WithOtel enables OpenTelemetry tracing for publisher, subscriber, and commit events.
//
// It injects a default Otel tracer implementation into the broker.
// Make sure your application is already setting up OpenTelemetry globally (e.g., exporter, tracer provider, etc).
func WithOtel() Options {
	return func(cfg *broker) {
		o := NewOtel()
		cfg.pubTracer = o
		cfg.subTracer = o
		cfg.commitTracer = o
	}
}

// WithYourCustomTracer allows you to inject your own custom tracer implementation
// for publishing, subscribing, and message committing.
//
// This is useful if you want to plug in a different tracing backend or a
// mock tracer for testing.
//
// Parameters:
//   - pub: Tracer for publish operations
//   - sub: Tracer for subscribe operations
//   - commit: Tracer for message commit operations
func WithYourCustomTracer(pub TracerPub, sub TracerSub, commit TracerCommitMessage) Options {
	return func(cfg *broker) {
		cfg.pubTracer = pub
		cfg.subTracer = sub
		cfg.commitTracer = commit
	}
}
