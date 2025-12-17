package otelsqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/databases/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName          = "github.com/SyaibanAhmadRamadhan/go-foundation-kit/databases/sqlx/otelsqlx"
	meterName           = "github.com/SyaibanAhmadRamadhan/go-foundation-kit/databases/sqlx/otelsqlx"
	sqlOperationUnknown = "UNKNOWN"
)

const (
	// RowsAffectedKey represents the number of rows affected.
	RowsAffectedKey = attribute.Key("sqlx.rows_affected")
	// QueryParametersKey represents the query parameters.
	QueryParametersKey = attribute.Key("sqlx.query.parameters")

	SqlxInTx        = attribute.Key("sqlx.in_tx")
	SqlxUsePrepared = attribute.Key("sqlx.use_prepared")

	// OperationTypeKey represents the pgx tracer operation type
	OperationTypeKey = attribute.Key("sqlx.operation.type")
	// DBClientOperationErrorsKey represents the count of operation errors
	DBClientOperationErrorsKey = attribute.Key("db.client.operation.errors")
)

// Tracer is a wrapper around the pgx tracer interfaces which instrument
// queries with both tracing and metrics.
type Tracer struct {
	tracer      trace.Tracer
	meter       metric.Meter
	tracerAttrs []attribute.KeyValue
	meterAttrs  []attribute.KeyValue

	operationDuration metric.Int64Histogram
	operationErrors   metric.Int64Counter

	trimQuerySpanName   bool
	spanNameFunc        SpanNameFunc
	prefixQuerySpanName bool
	logSQLStatement     bool
	includeParams       bool
}

type tracerConfig struct {
	tracerProvider trace.TracerProvider
	meterProvider  metric.MeterProvider

	tracerAttrs []attribute.KeyValue
	meterAttrs  []attribute.KeyValue

	trimQuerySpanName   bool
	spanNameFunc        SpanNameFunc
	prefixQuerySpanName bool
	logSQLStatement     bool
	includeParams       bool
}

// NewTracer returns a new Tracer.
func NewTracer(opts ...Option) *Tracer {
	cfg := &tracerConfig{
		tracerProvider:      otel.GetTracerProvider(),
		meterProvider:       otel.GetMeterProvider(),
		tracerAttrs:         []attribute.KeyValue{},
		meterAttrs:          []attribute.KeyValue{},
		trimQuerySpanName:   false,
		spanNameFunc:        nil,
		prefixQuerySpanName: true,
		logSQLStatement:     true,
		includeParams:       false,
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	tracer := &Tracer{
		tracer:              cfg.tracerProvider.Tracer(tracerName, trace.WithInstrumentationVersion(findOwnImportedVersion())),
		meter:               cfg.meterProvider.Meter(meterName, metric.WithInstrumentationVersion(findOwnImportedVersion())),
		tracerAttrs:         cfg.tracerAttrs,
		meterAttrs:          cfg.meterAttrs,
		trimQuerySpanName:   cfg.trimQuerySpanName,
		spanNameFunc:        cfg.spanNameFunc,
		prefixQuerySpanName: cfg.prefixQuerySpanName,
		logSQLStatement:     cfg.logSQLStatement,
		includeParams:       cfg.includeParams,
	}

	tracer.createMetrics()

	return tracer
}

func findOwnImportedVersion() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		for _, dep := range buildInfo.Deps {
			if dep.Path == tracerName {
				return dep.Version
			}
		}
	}

	return "unknown"
}

// recordSpanError handles all error handling to be applied on the provided span.
// The provided error must be non-nil and not a sql.ErrNoRows error.
// Otherwise, recordSpanError will be a no-op.
func recordSpanError(span trace.Span, err error) {
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

func (t *Tracer) Before(ctx context.Context, info *sqlx.HookInfo) context.Context {
	info.Start = time.Now()
	if !trace.SpanFromContext(ctx).IsRecording() {
		return ctx
	}

	opts := make([]trace.SpanStartOption, 0, 6)
	opts = append(opts,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(t.tracerAttrs...),
		trace.WithAttributes(
			SqlxInTx.Bool(info.InTx),
			SqlxUsePrepared.Bool(info.Prepared),
		),
	)

	if t.logSQLStatement {
		opts = append(opts, trace.WithAttributes(
			semconv.DBQueryText(info.SQL),
			semconv.DBOperationName(t.sqlOperationName(info.SQL)),
		))

		if t.includeParams {
			opts = append(opts, trace.WithAttributes(makeParamsAttribute(info.Args)))
		}
	}

	spanName := info.SQL
	if t.trimQuerySpanName {
		spanName = t.sqlOperationName(info.SQL)
	}
	spanName = fmt.Sprintf("%s - %s", info.Op, spanName)

	if t.prefixQuerySpanName {
		spanName = "query " + spanName
	}

	ctx, _ = t.tracer.Start(ctx, spanName, opts...)

	return ctx
}

func (t *Tracer) After(ctx context.Context, info *sqlx.HookInfo) {
	span := trace.SpanFromContext(ctx)
	recordSpanError(span, info.Err)
	t.incrementOperationErrorCount(ctx, info.Err, string(info.Op))

	if info.Rows != nil {
		span.SetAttributes(
			RowsAffectedKey.Int64(*info.Rows),
		)
	}

	span.End()

	t.recordOperationDuration(ctx, string(info.Op), info.Start)
}

func makeParamsAttribute(args []any) attribute.KeyValue {
	ss := make([]string, len(args))
	for i := range args {
		ss[i] = fmt.Sprintf("%+v", args[i])
	}

	return QueryParametersKey.StringSlice(ss)
}
