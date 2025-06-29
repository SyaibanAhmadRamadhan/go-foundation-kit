package libresty

import (
	"fmt"

	"go.opentelemetry.io/otel/trace"
	"resty.dev/v3"
)

type Option func(*restyTracing)

func WithCustomFormatter(fn spanNameFormatter) Option {
	return func(rt *restyTracing) {
		if fn != nil {
			rt.spanNameFormatter = fn
		}
	}
}

func WithCustomTracerName(name string) Option {
	return func(rt *restyTracing) {
		if name != "" {
			rt.tracer = rt.tracerProvider.Tracer(name)
		}
	}
}

func WithSpanStartOptions(opts ...trace.SpanStartOption) Option {
	return func(rt *restyTracing) {
		rt.spanOptions = opts
	}
}

type spanNameFormatter func(*resty.Response, *resty.Request) string

func defaultSpanNameFormatter(_ *resty.Response, req *resty.Request) string {
	return fmt.Sprintf("HTTP %s %s", req.Method, req.URL)
}
