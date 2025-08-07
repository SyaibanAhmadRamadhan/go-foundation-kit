package otelkafkax

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Option interface {
	apply(*openTelemetryConfig)
}

type optionFunc func(*openTelemetryConfig)

func (o optionFunc) apply(c *openTelemetryConfig) {
	o(c)
}

func WithTracerProvider(provider trace.TracerProvider) Option {
	return optionFunc(func(otc *openTelemetryConfig) {
		otc.tracerProvider = provider
	})
}

func WithPropagationProvider(propagation propagation.TextMapPropagator) Option {
	return optionFunc(func(otc *openTelemetryConfig) {
		otc.propagators = propagation
	})
}

func WithMeterProvider(provider metric.MeterProvider) Option {
	return optionFunc(func(otc *openTelemetryConfig) {
		otc.meterProvider = provider
	})
}

func WithTraceAttributes(attrs ...attribute.KeyValue) Option {
	return optionFunc(func(otc *openTelemetryConfig) {
		otc.tracerAttrs = append(otc.tracerAttrs, attrs...)
	})
}

func WithMeterAttributes(attrs ...attribute.KeyValue) Option {
	return optionFunc(func(otc *openTelemetryConfig) {
		otc.meterAttrs = append(otc.meterAttrs, attrs...)
	})
}
