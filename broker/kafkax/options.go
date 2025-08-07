package kafkax

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

// WithTracer sets custom tracer implementations for publish, subscribe, and commit operations.
//
// This can be useful when integrating with a different tracing backend
// or injecting mock tracers during testing.
//
// Parameters:
//   - pub: Tracer used for publishing messages.
//   - sub: Tracer used for subscribing to messages.
//   - commit: Tracer used for committing message processing.
func WithTracer(pub KafkaTracerPub, consume KafkaTracerConsume, commit KafkaTracerCommitMessage) Options {
	return func(cfg *broker) {
		cfg.pubTracer = pub
		cfg.consumeTracer = consume
		cfg.commitTracer = commit
	}
}
