package otelx

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// metricBuilder provides a fluent API for constructing and initializing
// an OpenTelemetry MeterProvider with various configurations.
// Its objective is to simplify the setup of metric collection components like exporters, readers,
// and views, ensuring metrics are correctly collected and exported to the desired backend.
type metricBuilder struct {
	exporter           sdkmetric.Exporter
	resource           *resource.Resource
	reader             sdkmetric.Reader
	view               sdkmetric.View
	periodicReaderOpts []sdkmetric.PeriodicReaderOption

	setGlobalMp bool
	initErr     error
}

// NewMetric creates a new metricBuilder instance.
// The objective is to provide a clean entry point for configuring OpenTelemetry metrics.
func NewMetric() *metricBuilder {
	return &metricBuilder{}
}

// WithExporter sets a pre-initialized OTLP metric exporter.
// The objective is to allow the user to provide a custom-configured OTLP exporter
// for sending metric data to a backend (e.g., Prometheus, OpenTelemetry Collector).
func (m *metricBuilder) WithExporter(e sdkmetric.Exporter) *metricBuilder {
	m.exporter = e
	return m
}

// WithResource sets a custom OpenTelemetry resource for the MeterProvider.
// The objective is to define the identity of the service generating the metrics.
func (m *metricBuilder) WithResource(r *resource.Resource) *metricBuilder {
	m.resource = r
	return m
}

// WithReader sets a custom metric Reader (e.g., ManualReader or PeriodicReader).
// The objective is to control how and when metrics are read from the SDK (e.g., pull-based vs push-based).
func (m *metricBuilder) WithReader(r sdkmetric.Reader) *metricBuilder {
	m.reader = r
	return m
}

// WithView sets a custom View for configuring metric instruments and aggregations.
// The objective is to allow fine-grained control over how metrics are processed,
// such as changing aggregation types or filtering which metrics are exported.
func (m *metricBuilder) WithView(v sdkmetric.View) *metricBuilder {
	m.view = v
	return m
}

// WithGlobalMetricProvider registers the created MeterProvider as the global default.
// The objective is to allow the application to use global OpenTelemetry metric helpers
// and ensure all libraries using the global provider report to this instance.
func (m *metricBuilder) WithGlobalMetricProvider() *metricBuilder {
	m.setGlobalMp = true
	return m
}

// Exporter returns the underlying exporter instance, useful for shutdown handling.
// The objective is to provide access to the exporter for manual management if needed.
func (m *metricBuilder) Exporter() sdkmetric.Exporter {
	return m.exporter
}

// WithPeriodicReader sets configuration options for the PeriodicReader.
// The objective is to allow tuning the frequency and timeout of metric exports
// when using the default periodic reader.
// These will be used only if no custom Reader is explicitly set.
func (m *metricBuilder) WithPeriodicReader(options ...sdkmetric.PeriodicReaderOption) *metricBuilder {
	m.periodicReaderOpts = append(m.periodicReaderOpts, options...)
	return m
}

// WithExporterGrpcBasicAuth initializes and sets an OTLP metric exporter using gRPC with Basic Auth.
// The objective is to provide a convenient way to connect to OTLP metric collectors
// that require Basic Authentication.
// Accepts optional otlpmetricgrpc.Options such as TLS or compression config.
func (m *metricBuilder) WithExporterGrpcBasicAuth(ctx context.Context, username, password, endpoint string, opts ...otlpmetricgrpc.Option) *metricBuilder {
	auth := fmt.Sprintf("%s:%s", username, password)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	basicAuth := fmt.Sprintf("Basic %s", encodedAuth)

	defaultOpts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithHeaders(map[string]string{
			"Authorization": basicAuth,
		}),
		otlpmetricgrpc.WithEndpoint(endpoint),
	}

	exp, err := otlpmetricgrpc.New(ctx, append(defaultOpts, opts...)...)
	if err != nil {
		m.initErr = fmt.Errorf("failed to create OTLP metric exporter: %w", err)
		return m
	}

	m.exporter = exp
	return m
}

// Init builds and initializes the MeterProvider using the provided configuration.
// Its objective is to finalize the metric setup, register the global provider if requested,
// and return a shutdown function for graceful cleanup.
// Returns a shutdown function and an error (if any).
func (m *metricBuilder) Init(ctx context.Context, serviceName string) (func(), error) {
	if m.initErr != nil {
		return nil, fmt.Errorf("otel meter builder initialization failed: %w", m.initErr)
	}

	var opts []sdkmetric.Option
	if m.reader != nil {
		opts = append(opts, sdkmetric.WithReader(m.reader))
	} else {
		if m.exporter != nil {
			reader := sdkmetric.NewPeriodicReader(m.exporter, m.periodicReaderOpts...)
			opts = append(opts, sdkmetric.WithReader(reader))
		} else {
			slog.Warn("no metric exporter configured; metric will be a no-op")
		}
	}

	// Resource setup
	if m.resource != nil {
		opts = append(opts, sdkmetric.WithResource(m.resource))
	} else {
		res, err := NewOpenTelemetryBasicResource(ctx, serviceName)
		if err != nil {
			return nil, fmt.Errorf("failed to create default OpenTelemetry resource: %w", err)
		}
		opts = append(opts, sdkmetric.WithResource(res))
	}

	if m.view != nil {
		opts = append(opts, sdkmetric.WithView(m.view))
	}

	mp := sdkmetric.NewMeterProvider(opts...)
	if m.setGlobalMp {
		otel.SetMeterProvider(mp)
	}

	shutdown := func() {
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		slog.Info("shutting down OpenTelemetry meter provider...")

		if err := mp.Shutdown(ctxShutdown); err != nil {
			slog.Error("failed to shutdown meter provider", slog.Any("error", err))
		} else {
			slog.Info("meter provider shutdown complete")
		}

		if m.exporter != nil {
			if err := m.exporter.Shutdown(ctxShutdown); err != nil {
				slog.Error("failed to shutdown meter exporter", slog.Any("error", err))
			} else {
				slog.Info("meter exporter shutdown complete")
			}
		}
	}

	return shutdown, nil
}
