package kafkax

import (
	"github.com/segmentio/kafka-go"
)

// Options defines a function that configures the broker during initialization.
// This is part of the functional options pattern commonly used in Go.
type Options func(cfg *broker)

// SetWriter configures a Kafka writer for the broker with the specified key, brokers, and topic.
//
// Parameters:
//   - key: identifier for the Kafka writer
//   - brokers: list of Kafka broker addresses
//   - topic: Kafka topic to which messages will be published
//
// Panics if no brokers are provided or if the topic is empty.
func SetWriter(key string, brokers []string, topic string) Options {
	return func(cfg *broker) {
		if len(brokers) == 0 {
			panic("kafka writer: no brokers provided")
		}
		if topic == "" {
			panic("kafka writer: topic cannot be empty")
		}
		cfg.writers[key] = &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		}
	}
}

// SetCustomWriter allows setting a pre-configured Kafka writer for the broker.
//
// This is useful when you need to customize the writer beyond the basic configuration
// provided by SetWriter.
//
// Parameters:
//   - key: identifier for the Kafka writer
//   - k: pre-configured *kafka.Writer instance
func SetCustomWriter(key string, k *kafka.Writer) Options {
	return func(cfg *broker) {
		cfg.writers[key] = k
	}
}

// WithTracer sets custom tracer implementations for publish, subscribe, and commit operations.
//
// This can be useful when integrating with a different tracing backend
// or injecting mock tracers during testing.
//
// Parameters:
//   - pub: Tracer used for publishing messages.
//   - sub: Tracer used for subscribing to messages.
//   - commit: Tracer used for committing message processing.
func WithTracer(pub TracerPub, consume TracerConsume, commit TracerCommitMessage) Options {
	return func(cfg *broker) {
		cfg.pubTracer = pub
		cfg.consumeTracer = consume
		cfg.commitTracer = commit
	}
}
