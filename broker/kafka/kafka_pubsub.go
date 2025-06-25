package libkafka

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/segmentio/kafka-go"
)

func (b *broker) Publish(ctx context.Context, input PubInput) (output PubOutput, err error) {
	if b.kafkaWriter == nil {
		return output, errors.New("kafka writer is not connected")
	}

	if len(input.Messages) <= 0 {
		return
	}

	var ctxTracer context.Context
	if b.pubTracer != nil {
		ctxTracer = b.pubTracer.TracePubStart(ctx, &input.Messages[0])
	}

	err = b.kafkaWriter.WriteMessages(ctx, input.Messages...)

	if b.pubTracer != nil {
		b.pubTracer.TracePubEnd(ctxTracer, output, err)
	}
	return
}

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
