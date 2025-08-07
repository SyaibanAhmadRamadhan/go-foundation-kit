package otelx

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// ExtractTraceparent injects the current trace context from the provided context into a carrier,
// and returns the "traceparent" header string.
//
// This is useful for propagating tracing info across services via HTTP headers, etc.
//
// Example return: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00"
func ExtractTraceparent(ctx context.Context) string {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	return carrier.Get("traceparent")
}

// ExtractTraceContext extracts traceparent and baggage headers from the context.
//
// Returns a map with keys like "traceparent", "baggage", and other custom headers if present.
func ExtractTraceContext(ctx context.Context) map[string]string {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	return carrier
}

// ExtractTraceparentAndBaggage returns the traceparent and baggage headers from the context.
func ExtractTraceparentAndBaggage(ctx context.Context) (traceparent string, baggage string) {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	return carrier.Get("traceparent"), carrier.Get("baggage")
}
