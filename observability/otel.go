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

func ExtractTraceparent(ctx context.Context) string {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	return carrier.Get("traceparent")
}
