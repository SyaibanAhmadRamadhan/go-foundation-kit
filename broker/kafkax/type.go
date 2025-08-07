//go:generate go tool mockgen -destination=../../.mocking/kafkax_mock.go -package=gofoundationkitmock . KafkaTracerPub,KafkaTracerConsume,KafkaTracerCommitMessage,KafkaPubSub

package kafkax

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// MarshalFunc defines a function that marshals any Go value into a byte slice.
// Commonly used for serializing Kafka messages before publishing.
type MarshalFunc func(any) ([]byte, error)

// UnmarshalFunc defines a function that unmarshals a byte slice into a Go value.
// Commonly used for deserializing Kafka message payloads when consuming.
type UnmarshalFunc func([]byte, any) error

// PubInput contains the messages to be published to Kafka.
type PubInput struct {
	Messages []kafka.Message
}

// PubOutput represents the result of a publish operation.
// Extend this struct if you want to return delivery metadata or message IDs.
type PubOutput struct{}

// SubInput defines the subscription input parameters.
// Includes Kafka reader configuration and an optional unmarshal function.
type SubInput struct {
	Config    kafka.ReaderConfig // Kafka reader configuration
	Unmarshal UnmarshalFunc      // Optional unmarshaller; defaults to JSON
}

// SubOutput contains the result of a subscription setup.
// Includes a Reader that can be used to fetch messages.
type SubOutput struct {
	Reader *Reader
}

// TracerPub is the interface for tracing publish operations.
// You can implement this to hook into publish start/end spans or logs.
type KafkaTracerPub interface {
	TracePubStart(ctx context.Context, msg *kafka.Message) context.Context
	TracePubEnd(ctx context.Context, input PubOutput, err error)
}

// TracerConsume is the interface for tracing Kafka message consumption.
// You can use it to track the lifecycle of a consumed message.
type KafkaTracerConsume interface {
	TraceConsumeStart(ctx context.Context, groupID string, msg *kafka.Message) context.Context
	TraceConsumeEnd(ctx context.Context, err error)
}

// TracerCommitMessage is the interface for tracing Kafka message commits.
// Useful for acknowledging offsets with observability support.
type KafkaTracerCommitMessage interface {
	TraceCommitMessagesStart(ctx context.Context, groupID string, messages ...kafka.Message) []context.Context
	TraceCommitMessagesEnd(ctx []context.Context, err error)
}

// PubSub defines a contract for a Kafka publisher-subscriber abstraction.
// It can be used to send and receive messages from Kafka with optional tracing and decoding.
type KafkaPubSub interface {
	Publish(ctx context.Context, input PubInput) (output PubOutput, err error)
	Subscribe(ctx context.Context, input SubInput) (output SubOutput, err error)
}
