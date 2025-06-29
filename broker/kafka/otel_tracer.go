package libkafka

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	// TracerName is the identifier used for this Kafka instrumentation.
	TracerName = "github.com/SyaibanAhmadRamadhan/go-foundation-kit/broker/kafka"

	// InstrumentVersion represents the semantic version of this instrumentation.
	InstrumentVersion = "v1.0.0"

	// kafkaLibName is the name of the underlying Kafka library.
	kafkaLibName = "github.com/segmentio/kafka-go"

	// netProtocolVer is the network protocol version used.
	netProtocolVer = "0.9.1"
)

var (
	// kafkaLibVersion will be populated dynamically from build info if available.
	kafkaLibVersion = "unknown"
)

// opentelemetry implements TracerPub, TracerSub, and TracerCommitMessage interfaces
// using OpenTelemetry for distributed tracing.
type opentelemetry struct {
	tracer      trace.Tracer
	propagators propagation.TextMapPropagator
	attrs       []attribute.KeyValue
	spanMap     map[uint64]trace.Span
}

// NewOtel initializes an OpenTelemetry tracer implementation for Kafka message operations.
//
// It sets the tracer name, instrumentation version, and default semantic attributes.
func NewOtel() *opentelemetry {
	tp := otel.GetTracerProvider()
	findOwnImportedVersion()
	return &opentelemetry{
		tracer:      tp.Tracer(TracerName, trace.WithInstrumentationVersion(InstrumentVersion)),
		propagators: otel.GetTextMapPropagator(),
		attrs: []attribute.KeyValue{
			semconv.ServiceName(kafkaLibName),
			semconv.ServiceVersion(kafkaLibVersion),
			semconv.MessagingSystemKafka,
			semconv.NetworkProtocolVersion(netProtocolVer),
			semconv.NetworkTransportTCP,
		},
		spanMap: make(map[uint64]trace.Span),
	}
}

// TracePubStart starts a span for a Kafka message being published.
// It injects the context into the Kafka message headers.
func (r *opentelemetry) TracePubStart(ctx context.Context, msg *kafka.Message) context.Context {
	carrier := NewMsgCarrier(msg)

	attrs := []attribute.KeyValue{
		semconv.MessagingKafkaMessageKey(string(msg.Key)),
		semconv.MessagingDestinationName(msg.Topic),
		semconv.MessagingOperationTypePublish,
		semconv.MessagingOperationName("send"),
		semconv.MessagingMessageBodySize(len(msg.Value)),
	}
	attrs = append(attrs, r.attrs...)

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
func (r *opentelemetry) TracePubEnd(ctx context.Context, input PubOutput, err error) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

// TraceSubStart starts a span for consuming a Kafka message.
// It extracts the context from message headers and sets relevant attributes.
func (r *opentelemetry) TraceSubStart(ctx context.Context, groupID string, msg *kafka.Message) context.Context {
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
	attrs = append(attrs, r.attrs...)

	opts := []trace.SpanStartOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindConsumer),
	}

	name := fmt.Sprintf("fetch from topic %s", msg.Topic)
	ctx, _ = r.tracer.Start(ctx, name, opts...)
	r.propagators.Inject(ctx, carrier)
	return ctx
}

// TraceSubEnd ends the span started for a Kafka message consumption.
// It records errors and sets status if unmarshalling or processing failed.
func (r *opentelemetry) TraceSubEnd(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, ErrJsonUnmarshal) {
			span.SetAttributes(semconv.ErrorTypeKey.String(ErrJsonUnmarshal.Error()))
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
		attrs = append(attrs, r.attrs...)

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
