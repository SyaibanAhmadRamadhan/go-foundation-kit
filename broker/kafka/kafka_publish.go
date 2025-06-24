package libkafka

import (
	"context"
	"errors"
)

func (b *broker) Publish(ctx context.Context, input PubInput) (output PubOutput, err error) {
	if b.kafkaWriter == nil {
		return output, errors.New("kafka writer is not connected")
	}

	if input.Messages == nil || len(input.Messages) <= 0 {
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
