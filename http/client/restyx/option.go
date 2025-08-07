package restyx

import (
	"fmt"

	"go.opentelemetry.io/otel/trace"
	"resty.dev/v3"
)

// Option defines a functional option for configuring Resty tracing.
type Option func(*restyTracing)

// WithCustomFormatter sets a custom span name formatter for HTTP requests.
// The formatter takes the Resty response and request and returns a span name string.
func WithCustomFormatter(fn spanNameFormatter) Option {
	return func(rt *restyTracing) {
		if fn != nil {
			rt.spanNameFormatter = fn
		}
	}
}

// WithCustomTracerName sets a custom OpenTelemetry tracer name.
// This is useful when you want to group spans under a specific instrumentation name.
func WithCustomTracerName(name string) Option {
	return func(rt *restyTracing) {
		if name != "" {
			rt.tracer = rt.tracerProvider.Tracer(name)
		}
	}
}

// WithSpanStartOptions allows you to inject custom span start options
// such as additional attributes, links, or custom sampling behavior.
func WithSpanStartOptions(opts ...trace.SpanStartOption) Option {
	return func(rt *restyTracing) {
		rt.spanOptions = opts
	}
}

// spanNameFormatter defines a function signature used to format span names
// based on the Resty request and response.
type spanNameFormatter func(*resty.Response, *resty.Request) string

// defaultSpanNameFormatter provides a basic formatter that uses the HTTP method and URL.
func defaultSpanNameFormatter(_ *resty.Response, req *resty.Request) string {
	return fmt.Sprintf("HTTP %s %s", req.Method, req.URL)
}
