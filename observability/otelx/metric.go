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
func NewMetric() *metricBuilder {
	return &metricBuilder{}
}

// WithExporter sets a pre-initialized OTLP metric exporter.
func (m *metricBuilder) WithExporter(e sdkmetric.Exporter) *metricBuilder {
	m.exporter = e
	return m
}

// WithResource sets a custom OpenTelemetry resource for the MeterProvider.
func (m *metricBuilder) WithResource(r *resource.Resource) *metricBuilder {
	m.resource = r
	return m
}

// WithReader sets a custom metric Reader (e.g., ManualReader or PeriodicReader).
func (m *metricBuilder) WithReader(r sdkmetric.Reader) *metricBuilder {
	m.reader = r
	return m
}

// WithView sets a custom View for configuring metric instruments and aggregations.
func (m *metricBuilder) WithView(v sdkmetric.View) *metricBuilder {
	m.view = v
	return m
}

// WithGlobalMetricProvider registers the created MeterProvider as the global default.
func (m *metricBuilder) WithGlobalMetricProvider() *metricBuilder {
	m.setGlobalMp = true
	return m
}

// Exporter returns the underlying exporter instance, useful for shutdown handling.
func (m *metricBuilder) Exporter() sdkmetric.Exporter {
	return m.exporter
}

// WithPeriodicReader sets configuration options for the PeriodicReader.
// These will be used only if no custom Reader is explicitly set.
func (m *metricBuilder) WithPeriodicReader(options ...sdkmetric.PeriodicReaderOption) *metricBuilder {
	m.periodicReaderOpts = append(m.periodicReaderOpts, options...)
	return m
}

// WithExporterGrpcBasicAuth initializes and sets an OTLP metric exporter using gRPC with Basic Auth.
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
		res, err := NewOpenTelemetryStdResource(ctx, serviceName)
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
