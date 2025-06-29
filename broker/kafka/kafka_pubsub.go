package libkafka

import (
	"context"
	"encoding/json"
	"errors"

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
	if b.kafkaWriter == nil {
		return output, errors.New("kafka writer is not connected")
	}

	if len(input.Messages) <= 0 {
		return
	}

	var ctxTracer []context.Context
	if b.pubTracer != nil {
		for _, v := range input.Messages {
			ctxTracer = append(ctxTracer, b.pubTracer.TracePubStart(ctx, &v))
		}
	}

	err = b.kafkaWriter.WriteMessages(ctx, input.Messages...)

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
	if err := PingKafkaBrokers(ctx, input.Config.Brokers, input.Config.Dialer); err != nil {
		return SubOutput{}, errors.New("unable to connect to kafka brokers: " + err.Error())
	}

	reader := kafka.NewReader(input.Config)
	if input.Unmarshal == nil {
		input.Unmarshal = json.Unmarshal
	}

	readerWrapper := &Reader{
		R:            reader,
		subTracer:    b.subTracer,
		commitTracer: b.commitTracer,
		groupID:      input.Config.GroupID,
		unmarshal:    input.Unmarshal,
	}

	b.readers = append(b.readers, readerWrapper)

	output = SubOutput{
		Reader: readerWrapper,
	}
	return
}
