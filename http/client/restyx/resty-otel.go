package restyx

import (
	"net/url"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"resty.dev/v3"
)

// restyTracing provides OpenTelemetry instrumentation for the Resty HTTP client.
type restyTracing struct {
	client            *resty.Client
	tracerProvider    trace.TracerProvider
	propagators       propagation.TextMapPropagator
	spanOptions       []trace.SpanStartOption
	tracer            trace.Tracer
	spanNameFormatter spanNameFormatter
}

// NewOtel enables OpenTelemetry tracing for the provided Resty client.
// It injects trace context into outbound requests, creates spans, and records
// HTTP metadata for distributed tracing.
func NewOtel(client *resty.Client, opts ...Option) {
	client = client.EnableTrace()
	tracing := &restyTracing{
		client:            client,
		tracerProvider:    otel.GetTracerProvider(),
		propagators:       otel.GetTextMapPropagator(),
		spanOptions:       []trace.SpanStartOption{},
		tracer:            otel.Tracer("github.com/SyaibanAhmadRamadhan/go-foundation-kit/httpclient/resty"),
		spanNameFormatter: defaultSpanNameFormatter,
	}

	for _, opt := range opts {
		opt(tracing)
	}

	client.SetRequestMiddlewares(
		tracing.onBeforeRequest(),
		resty.PrepareRequestMiddleware,
	)
	client.SetResponseMiddlewares(
		tracing.onAfterResponse(),
	)
	client.OnError(tracing.onError())
}

// onBeforeRequest creates a Resty middleware to start a tracing span and inject headers before the request is sent.
func (rt *restyTracing) onBeforeRequest() resty.RequestMiddleware {
	return func(_ *resty.Client, req *resty.Request) error {
		ctx, _ := rt.tracer.Start(req.Context(), rt.spanNameFormatter(nil, req), rt.spanOptions...)
		rt.propagators.Inject(ctx, propagation.HeaderCarrier(req.Header))
		req.SetContext(ctx)
		return nil
	}
}

// onAfterResponse creates a Resty middleware to record HTTP response metadata and finish the tracing span.
func (rt *restyTracing) onAfterResponse() resty.ResponseMiddleware {
	return func(_ *resty.Client, res *resty.Response) error {
		span := trace.SpanFromContext(res.Request.Context())
		if !span.IsRecording() {
			return nil
		}

		req := res.Request
		u, _ := url.Parse(req.URL)

		// Add HTTP standard attributes
		span.SetAttributes(
			attribute.String("http.method", req.Method),
			attribute.String("http.url", req.URL),
			attribute.String("http.scheme", u.Scheme),
			attribute.String("http.host", u.Host),
			attribute.String("http.path", u.Path),
			attribute.Int("http.status_code", res.StatusCode()),
			attribute.String("http.status_text", res.Status()),
		)

		// Add timing attributes from Resty TraceInfo
		ti := req.TraceInfo()
		span.SetAttributes(
			attribute.String("otel.trace.dns_lookup", ti.DNSLookup.String()),
			attribute.String("otel.trace.tcp_connect", ti.TCPConnTime.String()),
			attribute.String("otel.trace.tls_handshake", ti.TLSHandshake.String()),
			attribute.String("otel.trace.server_time", ti.ServerTime.String()),
			attribute.String("otel.trace.response_time", ti.ResponseTime.String()),
			attribute.String("otel.trace.total_time", ti.TotalTime.String()),
			attribute.Bool("otel.trace.is_conn_reused", ti.IsConnReused),
			attribute.Bool("otel.trace.is_conn_was_idle", ti.IsConnWasIdle),
			attribute.String("otel.trace.conn_idle_time", ti.ConnIdleTime.String()),
			attribute.Int("otel.trace.request_attempt", ti.RequestAttempt),
		)

		span.SetStatus(httpStatusToOtel(res.StatusCode()))
		span.SetName(rt.spanNameFormatter(res, req))
		span.End()
		return nil
	}
}

// onError returns a Resty error hook to record and finish a span if a request fails.
func (r *restyTracing) onError() resty.ErrorHook {
	return func(req *resty.Request, err error) {
		span := trace.SpanFromContext(req.Context())
		if !span.IsRecording() {
			return
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.SetName(r.spanNameFormatter(nil, req))
		span.End()
	}
}

// httpStatusToOtel maps HTTP status codes to OpenTelemetry codes.
func httpStatusToOtel(code int) (codes.Code, string) {
	if code >= 100 && code < 400 {
		return codes.Unset, "Successfully http client"
	}
	return codes.Error, "Failed http client"
}
