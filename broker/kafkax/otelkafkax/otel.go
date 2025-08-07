package otelkafkax

import (
	"runtime/debug"
	"slices"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type ctxKey string

const (
	// TracerName is the identifier used for this Kafka instrumentation.
	TracerName = "github.com/SyaibanAhmadRamadhan/go-foundation-kit/broker/kafkax"

	// InstrumentVersion represents the semantic version of this instrumentation.
	InstrumentVersion = "v1.0.0"

	// kafkaLibName is the name of the underlying Kafka library.
	kafkaLibName = "github.com/segmentio/kafka-go"

	// netProtocolVer is the network protocol version used.
	netProtocolVer = "0.9.1"

	startTimeCtxKey ctxKey = "otelKafkaxStartTime"
	topicKey        ctxKey = "kafkaTopicOtel"
)

var (
	// kafkaLibVersion will be populated dynamically from build info if available.
	kafkaLibVersion = "unknown"
)

// opentelemetry implements TracerPub, TracerSub, and TracerCommitMessage interfaces
// using OpenTelemetry for distributed tracing.
type opentelemetry struct {
	tracer      trace.Tracer
	meter       metric.Meter
	propagators propagation.TextMapPropagator

	msgPublishCount   metric.Int64Counter
	publishLatency    metric.Int64Histogram
	errorPublishCount metric.Int64Counter

	msgConsumeCount   metric.Int64Counter
	consumeLatency    metric.Int64Histogram
	errorConsumeCount metric.Int64Counter

	traceAttrs []attribute.KeyValue
	meterAttrs []attribute.KeyValue
	spanMap    map[uint64]trace.Span
}

type openTelemetryConfig struct {
	tracerProvider trace.TracerProvider
	meterProvider  metric.MeterProvider
	propagators    propagation.TextMapPropagator

	tracerAttrs []attribute.KeyValue
	meterAttrs  []attribute.KeyValue
}

// NewOtel initializes an OpenTelemetry tracer implementation for Kafka message operations.
//
// It sets the tracer name, instrumentation version, and default semantic attributes.
func NewOtel(opts ...Option) *opentelemetry {
	findOwnImportedVersion()

	otelConfig := &openTelemetryConfig{
		tracerProvider: otel.GetTracerProvider(),
		meterProvider:  otel.GetMeterProvider(),
		propagators:    otel.GetTextMapPropagator(),
		tracerAttrs: []attribute.KeyValue{
			semconv.ServiceName(kafkaLibName),
			semconv.ServiceVersion(kafkaLibVersion),
			semconv.MessagingSystemKafka,
			semconv.NetworkProtocolVersion(netProtocolVer),
			semconv.NetworkTransportTCP,
		},
		meterAttrs: []attribute.KeyValue{
			semconv.ServiceName(kafkaLibName),
			semconv.ServiceVersion(kafkaLibVersion),
			semconv.MessagingSystemKafka,
			semconv.NetworkProtocolVersion(netProtocolVer),
			semconv.NetworkTransportTCP,
		},
	}

	for _, opt := range opts {
		opt.apply(otelConfig)
	}

	o := &opentelemetry{
		tracer:      otelConfig.tracerProvider.Tracer(TracerName, trace.WithInstrumentationVersion(InstrumentVersion)),
		meter:       otelConfig.meterProvider.Meter(TracerName, metric.WithInstrumentationVersion(InstrumentVersion)),
		propagators: otelConfig.propagators,
		traceAttrs:  otelConfig.tracerAttrs,
		meterAttrs:  otelConfig.meterAttrs,
		spanMap:     make(map[uint64]trace.Span),
	}
	o.createMetrics()
	return o
}

// textMapCarrier is an implementation of OpenTelemetry's TextMapCarrier interface
// for injecting and extracting trace context from Kafka message headers.
type textMapCarrier struct {
	msg *kafka.Message
}

// Ensure that textMapCarrier implements the propagation.TextMapCarrier interface.
var _ propagation.TextMapCarrier = (*textMapCarrier)(nil)

// NewMsgCarrier returns a new textMapCarrier that wraps the given Kafka message.
//
// This is typically used for propagating OpenTelemetry trace context through
// Kafka message headers.
func NewMsgCarrier(msg *kafka.Message) *textMapCarrier {
	return &textMapCarrier{msg}
}

// Get retrieves the value associated with the given key from the Kafka message headers.
// If the key does not exist, an empty string is returned.
func (c *textMapCarrier) Get(key string) string {
	for _, h := range c.msg.Headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

// Set sets or replaces the value for the given key in the Kafka message headers.
// If the key already exists, it will be removed and replaced with the new value.
func (c *textMapCarrier) Set(key, value string) {
	// Ensure the uniqueness of the key by deleting existing headers with the same key.
	for i := len(c.msg.Headers) - 1; i >= 0; i-- {
		if c.msg.Headers[i].Key == key {
			c.msg.Headers = slices.Delete(c.msg.Headers, i, i+1)
		}
	}
	c.msg.Headers = append(c.msg.Headers, kafka.Header{
		Key:   key,
		Value: []byte(value),
	})
}

// Keys returns a list of all header keys currently stored in the Kafka message.
func (c *textMapCarrier) Keys() []string {
	out := make([]string, len(c.msg.Headers))
	for i, h := range c.msg.Headers {
		out[i] = h.Key
	}
	return out
}

// findOwnImportedVersion reads the build info to detect and store the imported version of the Tracer library.
// This is useful for debugging or version diagnostics in observability systems.
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
