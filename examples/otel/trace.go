package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/observability/otelx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	ctx := context.Background()

	// Custom Sampler: 10% sampling rate
	sampler := trace.ParentBased(trace.TraceIDRatioBased(0.1))

	// Custom SpanLimits (optional)
	spanLimits := trace.NewSpanLimits()
	spanLimits.AttributeValueLengthLimit = 256
	spanLimits.AttributeCountLimit = 32

	// Optional: custom propagator (adds custom baggage keys, for example)
	customPropagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	// Optional: add custom span processor (like simple processor)
	// You could also create your own implementation of SpanProcessor
	// but here we use SimpleSpanProcessor just as demo
	// simpleProcessor := trace.NewSimpleSpanProcessor(exporter)

	// Init tracing with chaining builder
	shutdown, err := otelx.NewTrace().
		WithSampler(sampler).
		WithSpanLimits(spanLimits).
		WithPropagator(customPropagator).
		WithIDGenerator(nil).
		WithExporterGrpcBasicAuth(
			ctx,
			"my-username",
			"my-password",
			"otel-collector.example.com:4317",
		).
		WithGlobalTraceProvider().
		Init(ctx, "awesome-app")
	if err != nil {
		log.Fatalf("failed to initialize OpenTelemetry tracing: %v", err)
	}
	defer shutdown()

	runAppLogic(ctx)
}

func runAppLogic(ctx context.Context) {
	tracer := otel.Tracer("awesome-app/component")

	ctx, span := tracer.Start(ctx, "runAppLogic")
	defer span.End()

	span.SetAttributes(
		attribute.String("env", os.Getenv("ENV")),
		attribute.String("component", "main-run"),
	)

	doSomething(ctx)
}

func doSomething(ctx context.Context) {
	tracer := otel.Tracer("awesome-app/component")

	_, span := tracer.Start(ctx, "doSomething")
	defer span.End()

	span.AddEvent("start doing work")

	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)

	span.AddEvent("work done")
}
