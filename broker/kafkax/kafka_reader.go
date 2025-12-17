package kafkax

import (
	"context"
	"errors"

	"github.com/segmentio/kafka-go"
)

// Reader is a wrapper around kafka.Reader that adds support for tracing and JSON unmarshalling.
//
// It integrates with OpenTelemetry-compatible Tracer interfaces (KafkaTracerConsume and KafkaTracerCommitMessage)
// to trace subscription and commit events, and allows automatic unmarshalling of Kafka message values.
type Reader struct {
	R             *kafka.Reader       // The underlying kafka.Reader
	consumeTracer TracerConsume       // Optional tracer for subscribe lifecycle
	commitTracer  TracerCommitMessage // Optional tracer for commit lifecycle
	groupID       string              // Kafka consumer group ID
	unmarshal     UnmarshalFunc       // Function to unmarshal message value into provided struct
}

// traceAndUnmarshal performs tracing (if enabled) and attempts to unmarshal the Kafka message value
// into the given struct pointer `v`.
//
// Parameters:
//   - ctx: parent context
//   - msg: the Kafka message to process
//   - v: optional target struct for JSON unmarshalling
//
// Returns:
//   - traced context if tracing is enabled
//   - error if unmarshalling fails or tracing hooks return errors
func (r *Reader) traceAndUnmarshal(ctx context.Context, msg kafka.Message, v any) (context.Context, error) {
	var ctxOtel context.Context
	var err error
	if r.consumeTracer != nil {
		ctxOtel = r.consumeTracer.TraceConsumeStart(ctx, r.groupID, &msg)
	}
	if v != nil {
		err = r.unmarshal(msg.Value, v)
		if err != nil {
			err = errors.Join(ErrJsonUnmarshal, err)
		}
	}
	if r.consumeTracer != nil {
		r.consumeTracer.TraceConsumeEnd(ctxOtel, err)
	}
	return ctxOtel, err
}

// FetchMessage wraps kafka.Reader.FetchMessage and adds tracing and optional unmarshalling.
//
// This function is useful when you want to fetch messages manually (non-blocking).
func (r *Reader) FetchMessage(ctx context.Context, v any) (kafka.Message, error) {
	msg, err := r.R.FetchMessage(ctx)
	if err != nil {
		return kafka.Message{}, err
	}
	_, err = r.traceAndUnmarshal(ctx, msg, v)
	return msg, err
}

// ReadMessage wraps kafka.Reader.ReadMessage and adds tracing and optional unmarshalling.
//
// Unlike FetchMessage, this blocks until a message is received.
func (r *Reader) ReadMessage(ctx context.Context, v any) (kafka.Message, error) {
	msg, err := r.R.ReadMessage(ctx)
	if err != nil {
		return kafka.Message{}, err
	}
	_, err = r.traceAndUnmarshal(ctx, msg, v)
	return msg, err
}

// CommitMessages commits the specified Kafka messages and invokes tracing hooks if enabled.
//
// Parameters:
//   - ctx: context
//   - messages: list of messages to commit
//
// Returns:
//   - error if commit fails
func (r *Reader) CommitMessages(ctx context.Context, messages ...kafka.Message) error {
	if len(messages) <= 0 {
		return nil
	}

	contexts := make([]context.Context, 0)
	if r.commitTracer != nil {
		contexts = r.commitTracer.TraceCommitMessagesStart(ctx, r.groupID, messages...)
	}

	err := r.R.CommitMessages(ctx, messages...)
	if r.commitTracer != nil {
		r.commitTracer.TraceCommitMessagesEnd(contexts, err)
	}

	return err
}

// Close closes the underlying kafka.Reader instance.
func (r *Reader) Close() error {
	return r.R.Close()
}

func (r *Reader) KafkaReader() *kafka.Reader {
	return r.R
}
