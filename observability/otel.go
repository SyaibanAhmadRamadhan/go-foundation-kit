package observability

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
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
	log.Printf("info: initialization opentelemetry successfully")
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
		err = fmt.Errorf("failed to created resource: %w", err)
		return nil, nil, err
	}

	bsp := trace.NewBatchSpanProcessor(exporter)

	provider := trace.NewTracerProvider(
		trace.WithSpanProcessor(bsp),
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
	)

	closeFn := func() {
		ctxClosure, cancelClosure := context.WithTimeout(ctx, 5*time.Second)
		defer cancelClosure()

		if err := exporter.Shutdown(ctxClosure); err != nil {
			log.Printf("error: failed to shutdown exporter: %v", err)
		}

		if err := provider.Shutdown(ctxClosure); err != nil {
			log.Printf("error: failed to shutdown provider: %v", err)
		}

		log.Printf("info: shutdown export and provider successfully")
	}

	return provider, closeFn, nil
}

func ExtractTraceparent(ctx context.Context) string {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	return carrier.Get("traceparent")
}
