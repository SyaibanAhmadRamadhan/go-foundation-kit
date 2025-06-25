package libkafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type MarshalFunc func(any) ([]byte, error)
type UnmarshalFunc func([]byte, any) error

type PubInput struct {
	Messages []kafka.Message
}

type PubOutput struct{}

type SubInput struct {
	Config kafka.ReaderConfig

	// by default using json
	Unmarshal UnmarshalFunc
}

type SubOutput struct {
	Reader *Reader
}

type TracerPub interface {
	TracePubStart(ctx context.Context, msg *kafka.Message) context.Context
	TracePubEnd(ctx context.Context, input PubOutput, err error)
}

type TracerSub interface {
	TraceSubStart(ctx context.Context, groupID string, msg *kafka.Message) context.Context
	TraceSubEnd(ctx context.Context, err error)
}

type TracerCommitMessage interface {
	TraceCommitMessagesStart(ctx context.Context, groupID string, messages ...kafka.Message) []context.Context
	TraceCommitMessagesEnd(ctx []context.Context, err error)
}

type PubSub interface {
	Publish(ctx context.Context, input PubInput) (output PubOutput, err error)
	Subscribe(ctx context.Context, input SubInput) (output SubOutput, err error)
}
