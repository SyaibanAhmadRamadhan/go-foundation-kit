package observability

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// NewOtel initializes OpenTelemetry tracing with OTLP exporter using basic authentication.
//
// Parameters:
//   - serviceName: the name of the service used for resource identification.
//   - otlpEndpoint: the OTLP gRPC endpoint (e.g., "localhost:4317").
//   - otlpUsername: username for Basic Auth.
//   - otlpPassword: password for Basic Auth.
//
// Returns:
//   - func(): a function to shut down the exporter and tracer provider gracefully.
//   - error: any error that occurs during setup.
func NewOtel(serviceName, otlpEndpoint, otlpUsername, otlpPassword string) (func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	auth := otlpUsername + ":" + otlpPassword
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	basicAuth := fmt.Sprintf("Basic %s", encodedAuth)

	traceClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithHeaders(map[string]string{
			"Authorization": basicAuth,
		}),
		otlptracegrpc.WithEndpoint(otlpEndpoint),
	)

	traceExp, err := otlptrace.New(ctx, traceClient)
	if err != nil {
		return nil, err
	}

	traceProvider, closeFunc, err := startTraceProvider(traceExp, serviceName)
	if err != nil {
		return nil, err
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(traceProvider)
	slog.Info("OpenTelemetry tracer provider initialized",
		slog.String("service", serviceName),
	)
	return closeFunc, nil
}

// startTraceProvider sets up the tracer provider with resource attributes and span processor.
//
// Parameters:
//   - exporter: the OTLP trace exporter.
//   - serviceName: the name of the service.
//
// Returns:
//   - *trace.TracerProvider: the initialized tracer provider.
//   - func(): a cleanup function to shut down the provider and exporter.
//   - error: any error during initialization.
func startTraceProvider(exporter *otlptrace.Exporter, serviceName string) (*trace.TracerProvider, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
		resource.WithHost(),
		resource.WithTelemetrySDK(),
		resource.WithFromEnv(),
	)
	if err != nil {
		slog.Error("failed to create OpenTelemetry resource",
			slog.String("service", serviceName),
			slog.Any("error", err),
		)
		return nil, nil, fmt.Errorf("failed to create OpenTelemetry resource: %w", err)
	}

	bsp := trace.NewBatchSpanProcessor(exporter)
	provider := trace.NewTracerProvider(
		trace.WithSpanProcessor(bsp),
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
	)

	closeFn := func() {
		ctxClosure, cancelClosure := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelClosure()

		slog.Info("shutting down OpenTelemetry components...")

		if err := exporter.Shutdown(ctxClosure); err != nil {
			slog.Error("failed to shutdown OpenTelemetry exporter", slog.Any("error", err))
		} else {
			slog.Info("OpenTelemetry exporter shutdown complete")
		}

		if err := provider.Shutdown(ctxClosure); err != nil {
			slog.Error("failed to shutdown tracer provider", slog.Any("error", err))
		} else {
			slog.Info("OpenTelemetry tracer provider shutdown complete")
		}
	}

	return provider, closeFn, nil
}

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
