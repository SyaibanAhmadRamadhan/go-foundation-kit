package otelx

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"time"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/utils/reflection"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// traceBuilder provides a fluent interface for constructing an OpenTelemetry TracerProvider.
// Its objective is to simplify the complex setup of tracing components like exporters, resources,
// samplers, and processors through a chainable API.
type traceBuilder struct {
	exporter      *otlptrace.Exporter
	resource      *resource.Resource
	sampler       sdktrace.Sampler
	idGenerator   sdktrace.IDGenerator
	spanProcessor sdktrace.SpanProcessor
	propagator    propagation.TextMapPropagator
	spanLimit     sdktrace.SpanLimits

	setGlobalTp bool
	initErr     error
}

// NewTrace returns a new instance of traceBuilder.
// The objective is to provide a clean entry point for configuring OpenTelemetry tracing.
func NewTrace() *traceBuilder {
	return &traceBuilder{}
}

// WithExporter sets the OTLP trace exporter.
// The objective is to allow the user to provide a custom-configured OTLP exporter
// for sending trace data to a backend (e.g., Jaeger, Tempo, or OpenTelemetry Collector).
func (b *traceBuilder) WithExporter(exp *otlptrace.Exporter) *traceBuilder {
	b.exporter = exp
	return b
}

// WithExporterGrpcBasicAuth initializes and sets an OTLP trace exporter over gRPC with basic auth.
// The objective is to provide a convenient way to connect to OTLP collectors that require
// Basic Authentication, handling the encoding and header setup automatically.
// You can pass additional otlptracegrpc.Option (e.g., WithTLS, WithCompressor, etc.).
// If an error occurs, it will be stored in the builder and returned during Init().
func (b *traceBuilder) WithExporterGrpcBasicAuth(ctx context.Context, username, password, endpoint string, opts ...otlptracegrpc.Option) *traceBuilder {
	auth := fmt.Sprintf("%s:%s", username, password)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	basicAuth := fmt.Sprintf("Basic %s", encodedAuth)

	defaultOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithHeaders(map[string]string{
			"Authorization": basicAuth,
		}),
		otlptracegrpc.WithEndpoint(endpoint),
	}

	traceClient := otlptracegrpc.NewClient(
		append(defaultOpts, opts...)...,
	)

	exp, err := otlptrace.New(ctx, traceClient)
	if err != nil {
		b.initErr = fmt.Errorf("failed to create OTLP trace exporter: %w", err)
		return b
	}

	b.exporter = exp
	return b
}

// WithResource sets the OpenTelemetry resource.
// The objective is to define the identity of the service being traced (e.g., name, version, environment).
// If not provided, a default resource with the given service name will be used.
func (b *traceBuilder) WithResource(res *resource.Resource) *traceBuilder {
	b.resource = res
	return b
}

// WithSampler sets a custom sampler.
// The objective is to control the volume of traces being collected and exported,
// which is important for managing performance and storage costs.
// Defaults to AlwaysSample if not specified.
func (b *traceBuilder) WithSampler(sampler sdktrace.Sampler) *traceBuilder {
	b.sampler = sampler
	return b
}

// WithPropagator sets the text map propagator for context injection and extraction.
// The objective is to enable distributed tracing by allowing the trace context to travel
// across service boundaries (e.g., via HTTP headers).
func (b *traceBuilder) WithPropagator(p propagation.TextMapPropagator) *traceBuilder {
	b.propagator = p
	return b
}

// WithSpanLimits sets the limits for spans (e.g., max attributes, max events).
// The objective is to prevent excessive resource consumption by individual spans
// that might have a large number of attributes or events.
func (b *traceBuilder) WithSpanLimits(sl sdktrace.SpanLimits) *traceBuilder {
	b.spanLimit = sl
	return b
}

// WithPropagators sets multiple propagators as a composite propagator.
// The objective is to support multiple propagation formats (e.g., W3C TraceContext and Baggage)
// in a single configuration.
func (b *traceBuilder) WithPropagators(ps ...propagation.TextMapPropagator) *traceBuilder {
	b.propagator = propagation.NewCompositeTextMapPropagator(ps...)
	return b
}

// WithIDGenerator sets a custom trace ID generator.
// The objective is to allow the use of specific ID generation logic if required
// by the tracing backend or organizational standards.
func (b *traceBuilder) WithIDGenerator(idGen sdktrace.IDGenerator) *traceBuilder {
	b.idGenerator = idGen
	return b
}

// WithSpanProcessor sets a custom span processor.
// The objective is to provide fine-grained control over how spans are processed
// (e.g., using a SimpleSpanProcessor for testing or a BatchSpanProcessor for production).
func (b *traceBuilder) WithSpanProcessor(sp sdktrace.SpanProcessor) *traceBuilder {
	b.spanProcessor = sp
	return b
}

// Exporter returns the currently configured OTLP trace exporter.
// The objective is to allow access to the exporter instance for inspection or further configuration.
func (b *traceBuilder) Exporter() *otlptrace.Exporter {
	return b.exporter
}

// WithGlobalTraceProvider flags that the configured TracerProvider should be set globally.
// The objective is to simplify usage by allowing the application to use otel.GetTracerProvider()
// and other global OpenTelemetry helpers.
func (b *traceBuilder) WithGlobalTraceProvider() *traceBuilder {
	b.setGlobalTp = true
	return b
}

// Init constructs the TracerProvider based on the builder configuration.
// Its objective is to finalize the setup, register the global provider if requested,
// and return a shutdown function for graceful cleanup.
// It sets the global TracerProvider if WithGlobalTraceProvider was called.
// Returns a shutdown function and any error that occurred during initialization.
func (b *traceBuilder) Init(ctx context.Context, serviceName string) (func(), error) {
	if b.initErr != nil {
		return nil, fmt.Errorf("otel trace builder initialization failed: %w", b.initErr)
	}

	var opts []sdktrace.TracerProviderOption

	// Span processor and exporter
	if b.spanProcessor != nil {
		opts = append(opts, sdktrace.WithSpanProcessor(b.spanProcessor))
	} else if b.exporter != nil {
		opts = append(opts, sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(b.exporter)))
	} else {
		slog.Warn("no span processor or exporter configured; tracing will be a no-op")
	}

	if b.exporter == nil && b.spanProcessor == nil {
		slog.Warn("no exporter or span processor; spans will not be exported")
	}

	// Resource setup
	if b.resource != nil {
		opts = append(opts, sdktrace.WithResource(b.resource))
	} else {
		res, err := NewOpenTelemetryBasicResource(ctx, serviceName)
		if err != nil {
			return nil, fmt.Errorf("failed to create default OpenTelemetry resource: %w", err)
		}
		opts = append(opts, sdktrace.WithResource(res))
	}

	// Sampler
	if b.sampler != nil {
		opts = append(opts, sdktrace.WithSampler(b.sampler))
	} else {
		opts = append(opts, sdktrace.WithSampler(sdktrace.AlwaysSample()))
	}

	// ID Generator
	if b.idGenerator != nil {
		opts = append(opts, sdktrace.WithIDGenerator(b.idGenerator))
	}

	if reflection.IsZero(b.spanLimit) {
		b.spanLimit = sdktrace.NewSpanLimits()
	}
	opts = append(opts, sdktrace.WithRawSpanLimits(b.spanLimit))

	// Build TracerProvider
	tp := sdktrace.NewTracerProvider(opts...)

	// Set global TracerProvider if configured
	if b.setGlobalTp {
		otel.SetTracerProvider(tp)
		slog.Info("OpenTelemetry tracer provider registered as global", slog.String("service", serviceName))
		if b.propagator != nil {
			otel.SetTextMapPropagator(b.propagator)
			slog.Info("OpenTelemetry propagator registered as global")
		} else {
			otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
			slog.Info("OpenTelemetry default propagator registered as global")
		}
	}

	// Return shutdown function
	shutdown := func() {
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		slog.Info("shutting down OpenTelemetry tracer provider...")

		if err := tp.Shutdown(ctxShutdown); err != nil {
			slog.Error("failed to shutdown tracer provider", slog.Any("error", err))
		} else {
			slog.Info("tracer provider shutdown complete")
		}

		if b.exporter != nil {
			if err := b.exporter.Shutdown(ctxShutdown); err != nil {
				slog.Error("failed to shutdown trace exporter", slog.Any("error", err))
			} else {
				slog.Info("trace exporter shutdown complete")
			}
		}
	}

	return shutdown, nil
}
