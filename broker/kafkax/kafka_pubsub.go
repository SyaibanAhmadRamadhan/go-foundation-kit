package kafkax

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/segmentio/kafka-go"
)

// Publish sends a batch of Kafka messages to the topic specified in each message.
//
// It optionally traces each message publish operation if a pubTracer is configured.
// The tracer's TracePubStart is called before publishing, and TracePubEnd is called after.
//
// Parameters:
//   - ctx: context for cancellation, timeout, and tracing
//   - input: PubInput containing the messages to be sent
//
// Returns:
//   - PubOutput (currently unused)
//   - error if kafkaWriter is nil or if WriteMessages fails
func (b *broker) Publish(ctx context.Context, input PubInput) (output PubOutput, err error) {
	writer, err := b.GetWriter(input.KeyWriter)
	if err != nil {
		return output, err
	}
	if writer == nil {
		return output, errors.New("kafka writer is not connected")
	}

	if len(input.Messages) <= 0 {
		return output, nil
	}

	var ctxTracer []context.Context
	if b.pubTracer != nil {
		for _, v := range input.Messages {
			ctxTracer = append(ctxTracer, b.pubTracer.TracePubStart(ctx, &v))
		}
	}

	err = writer.WriteMessages(ctx, input.Messages...)

	if b.pubTracer != nil {
		for _, v := range ctxTracer {
			b.pubTracer.TracePubEnd(v, output, err)
		}
	}
	return
}

// Subscribe initializes a new Kafka consumer and returns a Reader wrapper,
// which includes optional tracing and custom unmarshal function.
//
// Before subscribing, it pings the brokers to ensure connectivity.
//
// Parameters:
//   - ctx: context for timeout, cancellation
//   - input: SubInput containing kafka.ReaderConfig, custom Unmarshal function, etc.
//
// Returns:
//   - SubOutput containing a pointer to Reader
//   - error if brokers are unreachable or misconfigured
func (b *broker) Subscribe(ctx context.Context, input SubInput) (output SubOutput, err error) {
	if len(input.Config.Brokers) == 0 {
		return SubOutput{}, errors.New("no kafka brokers specified")
	}
	if input.Config.Topic == "" {
		return SubOutput{}, errors.New("kafka topic is required")
	}
	if input.Config.GroupID == "" {
		return SubOutput{}, errors.New("kafka group ID is required")
	}

	if err := PingBrokers(ctx, input.Config.Brokers, input.Config.Dialer); err != nil {
		return SubOutput{}, fmt.Errorf("unable to connect to kafka brokers: %w", err)
	}

	reader := kafka.NewReader(input.Config)
	if input.Unmarshal == nil {
		input.Unmarshal = json.Unmarshal
	}

	readerWrapper := &Reader{
		R:             reader,
		consumeTracer: b.consumeTracer,
		commitTracer:  b.commitTracer,
		groupID:       input.Config.GroupID,
		unmarshal:     input.Unmarshal,
	}

	b.readers = append(b.readers, readerWrapper)

	output = SubOutput{
		Reader: readerWrapper,
	}
	return
}
