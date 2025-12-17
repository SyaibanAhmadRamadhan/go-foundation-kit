package otelkafkax

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// createMetrics initializes all synchronous metrics tracked by Tracer.
// Any errors encountered upon metric creation will be sent to the globally assigned OpenTelemetry ErrorHandler.
func (o *opentelemetry) createMetrics() {
	msgPublishCount, err := o.meter.Int64Counter(
		"messaging.kafka.messages.produced",
		metric.WithDescription("Number of messages successfully produced to Kafka"),
	)
	if err != nil {
		otel.Handle(err)
	}

	publishLatency, err := o.meter.Int64Histogram(
		"messaging.kafka.produce.latency.ms",
		metric.WithDescription("Duration in milliseconds to produce Kafka messages"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		otel.Handle(err)
	}
	errorPublishCount, err := o.meter.Int64Counter(
		"messaging.kafka.publish.error.count",
		metric.WithDescription("Total number of Kafka publish errors"),
	)
	if err != nil {
		otel.Handle(err)
	}

	msgConsumeCount, err := o.meter.Int64Counter(
		"messaging.kafka.consume.count",
		metric.WithDescription("Number of Kafka messages successfully consumed"),
	)
	if err != nil {
		otel.Handle(err)
	}

	consumeLatency, err := o.meter.Int64Histogram(
		"messaging.kafka.consume.latency.ms",
		metric.WithUnit("ms"),
		metric.WithDescription("Latency in milliseconds to process a Kafka message from receipt to completion"),
	)
	if err != nil {
		otel.Handle(err)
	}

	errorConsumeCount, err := o.meter.Int64Counter(
		"messaging.kafka.consume.error.count",
		metric.WithDescription("Total number of Kafka consume errors"),
	)
	if err != nil {
		otel.Handle(err)
	}

	o.errorPublishCount = errorPublishCount
	o.msgPublishCount = msgPublishCount
	o.publishLatency = publishLatency
	o.consumeLatency = consumeLatency
	o.msgConsumeCount = msgConsumeCount
	o.errorConsumeCount = errorConsumeCount
}
