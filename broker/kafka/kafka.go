package libkafka

import (
	"context"
	"errors"
	"runtime/debug"

	"github.com/segmentio/kafka-go"
)

var ErrProcessShutdownIsRunning = errors.New("process shutdown is running")
var errClosed = errors.New("kafka closed")
var ErrJsonUnmarshal = errors.New("json unmarshal error")

type broker struct {
	kafkaWriter  *kafka.Writer
	pubTracer    TracerPub
	subTracer    TracerSub
	commitTracer TracerCommitMessage
}

func New(opts ...Options) *broker {
	b := &broker{}
	for _, option := range opts {
		option(b)
	}

	return b
}

func (b *broker) Close() {
	b.kafkaWriter.Close()
}

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
	Reader Reader
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
