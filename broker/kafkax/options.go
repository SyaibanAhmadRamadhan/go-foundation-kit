package kafkax

import (
	"github.com/segmentio/kafka-go"
)

// Options defines a function that configures the broker during initialization.
// This is part of the functional options pattern commonly used in Go.
type Options func(cfg *broker)

// Writer creates an option to initialize a Kafka writer using the given
// broker addresses and topic.
//
// Parameters:
//   - brokers: list of Kafka broker addresses (e.g., []string{"localhost:9092"})
//   - topic: the Kafka topic to which messages will be written
//
// The writer will use the LeastBytes balancer strategy to distribute messages.
func Writer(key string, brokers []string, topic string) Options {
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

// CustomWriter allows you to provide a fully customized *kafka.Writer instance.
//
// Useful if you want fine-grained control over Kafka writer configuration such as
// retries, batch size, compression, async, etc.
func CustomWriter(key string, k *kafka.Writer) Options {
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
