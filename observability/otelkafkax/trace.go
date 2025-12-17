package otelkafkax

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/broker/kafkax"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// TracePubStart starts a span for a Kafka message being published.
// It injects the context into the Kafka message headers.
func (r *opentelemetry) TracePubStart(ctx context.Context, msg *kafka.Message) context.Context {
	ctx = context.WithValue(ctx, startTimeCtxKey, time.Now())
	ctx = context.WithValue(ctx, topicKey, msg.Topic)
	carrier := NewMsgCarrier(msg)

	attrs := []attribute.KeyValue{
		semconv.MessagingKafkaMessageKey(string(msg.Key)),
		semconv.MessagingDestinationName(msg.Topic),
		semconv.MessagingOperationTypePublish,
		semconv.MessagingOperationName("send"),
		semconv.MessagingMessageBodySize(len(msg.Value)),
	}
	attrs = append(attrs, r.traceAttrs...)

	opts := []trace.SpanStartOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindProducer),
	}

	name := fmt.Sprintf("%s send", msg.Topic)
	ctx, _ = r.tracer.Start(ctx, name, opts...)
	r.propagators.Inject(ctx, carrier)
	return ctx
}

// TracePubEnd ends the span started for publishing a Kafka message.
// It records error details if provided.
func (r *opentelemetry) TracePubEnd(ctx context.Context, input kafkax.PubOutput, err error) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	topic := ctx.Value(topicKey)
	r.msgPublishCount.Add(ctx, 1, metric.WithAttributeSet(
		attribute.NewSet(append(r.meterAttrs, attribute.String("messaging.destination.name", fmt.Sprintf("%s", topic)))...),
	))
	if startTime, ok := ctx.Value(startTimeCtxKey).(time.Time); ok {
		r.publishLatency.Record(ctx, time.Since(startTime).Milliseconds(),
			metric.WithAttributeSet(attribute.NewSet(
				append(r.meterAttrs, attribute.String("messaging.destination.name", fmt.Sprintf("%s", topic)))...),
			))
	}

	if err != nil {
		r.errorPublishCount.Add(ctx, 1, metric.WithAttributeSet(
			attribute.NewSet(append(r.meterAttrs,
				attribute.String("messaging.destination.name", fmt.Sprintf("%s", topic)),
				attribute.String("error.type", fmt.Sprintf("%T", err)),
			)...),
		))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

// TraceConsumeStart starts a span for consuming a Kafka message.
// It extracts the context from message headers and sets relevant attributes.
func (r *opentelemetry) TraceConsumeStart(ctx context.Context, groupID string, msg *kafka.Message) context.Context {
	ctx = context.WithValue(ctx, startTimeCtxKey, time.Now())
	ctx = context.WithValue(ctx, topicKey, msg.Topic)

	carrier := NewMsgCarrier(msg)
	ctx = r.propagators.Extract(ctx, carrier)

	attrs := []attribute.KeyValue{
		semconv.MessagingKafkaMessageKey(string(msg.Key)),
		semconv.MessagingDestinationName(msg.Topic),
		semconv.MessagingOperationTypeReceive,
		semconv.MessagingOperationName("poll"),
		semconv.MessagingEventhubsConsumerGroup(groupID),
		semconv.MessagingDestinationPartitionID(strconv.FormatInt(int64(msg.Partition), 10)),
		semconv.MessagingMessageBodySize(len(msg.Value)),
		semconv.MessagingKafkaMessageOffset(int(msg.Offset)),
	}
	attrs = append(attrs, r.traceAttrs...)

	opts := []trace.SpanStartOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindConsumer),
	}

	name := fmt.Sprintf("fetch from topic %s", msg.Topic)
	ctx, _ = r.tracer.Start(ctx, name, opts...)
	r.propagators.Inject(ctx, carrier)
	return ctx
}

// TraceConsumeEnd ends the span started for a Kafka message consumption.
// It records errors and sets status if unmarshalling or processing failed.
func (r *opentelemetry) TraceConsumeEnd(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	topic := ctx.Value(topicKey)
	r.msgConsumeCount.Add(ctx, 1, metric.WithAttributeSet(
		attribute.NewSet(append(r.meterAttrs, attribute.String("messaging.destination.name", fmt.Sprintf("%s", topic)))...),
	))
	if startTime, ok := ctx.Value(startTimeCtxKey).(time.Time); ok {
		r.consumeLatency.Record(ctx, time.Since(startTime).Milliseconds(),
			metric.WithAttributeSet(attribute.NewSet(
				append(r.meterAttrs, attribute.String("messaging.destination.name", fmt.Sprintf("%s", topic)))...),
			))
	}

	if err != nil {
		r.errorConsumeCount.Add(ctx, 1, metric.WithAttributeSet(
			attribute.NewSet(append(r.meterAttrs,
				attribute.String("messaging.destination.name", fmt.Sprintf("%s", topic)),
				attribute.String("error.type", fmt.Sprintf("%T", err)),
			)...),
		))

		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, kafkax.ErrJsonUnmarshal) {
			span.SetAttributes(semconv.ErrorTypeKey.String(kafkax.ErrJsonUnmarshal.Error()))
		}
	}
	span.End()
}

// TraceCommitMessagesStart starts a span for each Kafka message being committed.
// This is useful for observing commit acknowledgment timing.
func (r *opentelemetry) TraceCommitMessagesStart(ctx context.Context, groupID string, messages ...kafka.Message) []context.Context {
	if messages == nil {
		return make([]context.Context, 0)
	}

	contexts := make([]context.Context, 0, len(messages))

	for _, msg := range messages {

		carrier := NewMsgCarrier(&msg)
		ctx = r.propagators.Extract(ctx, carrier)

		attrs := []attribute.KeyValue{
			semconv.MessagingKafkaMessageKey(string(msg.Key)),
			semconv.MessagingDestinationName(msg.Topic),
			semconv.MessagingOperationTypeSettle,
			semconv.MessagingOperationName("commit"),
			semconv.MessagingEventhubsConsumerGroup(groupID),
			semconv.MessagingDestinationPartitionID(strconv.FormatInt(int64(msg.Partition), 10)),
			semconv.MessagingMessageBodySize(len(msg.Value)),
			semconv.MessagingKafkaMessageOffset(int(msg.Offset)),
		}
		attrs = append(attrs, r.traceAttrs...)

		opts := []trace.SpanStartOption{
			trace.WithAttributes(attrs...),
			trace.WithSpanKind(trace.SpanKindConsumer),
		}

		name := fmt.Sprintf("commit from topic %s", msg.Topic)
		ctx, _ = r.tracer.Start(ctx, name, opts...)
		contexts = append(contexts, ctx)
	}

	return contexts
}

// TraceCommitMessagesEnd ends all spans started during message commit tracing.
// It also records errors if commit failed.
func (r *opentelemetry) TraceCommitMessagesEnd(ctxs []context.Context, err error) {
	for _, c := range ctxs {
		span := trace.SpanFromContext(c)
		if !span.IsRecording() {
			continue
		}
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			span.SetAttributes(semconv.ErrorTypeKey.String("failed commit message"))
		}
		span.End()
	}
}
