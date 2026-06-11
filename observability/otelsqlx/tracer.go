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
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
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

	// OperationTypeKey represents the sqlx tracer operation type
	OperationTypeKey = attribute.Key("sqlx.operation.type")
	// DBClientOperationErrorsKey represents the count of operation errors
	DBClientOperationErrorsKey = attribute.Key("db.client.operation.errors")
)

// Tracer is a wrapper around the sqlx hook interfaces which instrument
// queries with both tracing and metrics.
type Tracer struct {
	tracer      trace.Tracer
	meter       metric.Meter
	tracerAttrs []attribute.KeyValue
	meterAttrs  []attribute.KeyValue

	operationDuration metric.Int64Histogram
	operationErrors   metric.Int64Counter

	dbSystem            string
	dbNamespace         string
	serverAddress       string
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

	dbSystem            string
	dbNamespace         string
	serverAddress       string
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
		dbSystem:            "postgresql", // Default to postgres
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
		dbSystem:            cfg.dbSystem,
		dbNamespace:         cfg.dbNamespace,
		serverAddress:       cfg.serverAddress,
		trimQuerySpanName:   cfg.trimQuerySpanName,
		spanNameFunc:        cfg.spanNameFunc,
		prefixQuerySpanName: cfg.prefixQuerySpanName,
		logSQLStatement:     cfg.logSQLStatement,
		includeParams:       cfg.includeParams,
	}

	// Add database identity to metrics by default
	tracer.meterAttrs = append(tracer.meterAttrs, semconv.DBSystemKey.String(tracer.dbSystem))
	if tracer.dbNamespace != "" {
		tracer.meterAttrs = append(tracer.meterAttrs, semconv.DBNamespace(tracer.dbNamespace))
	}
	if tracer.serverAddress != "" {
		tracer.meterAttrs = append(tracer.meterAttrs, semconv.ServerAddress(tracer.serverAddress))
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

	opts := make([]trace.SpanStartOption, 0, 8)
	opts = append(opts,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(t.tracerAttrs...),
		trace.WithAttributes(
			SqlxInTx.Bool(info.InTx),
			SqlxUsePrepared.Bool(info.Prepared),
		),
	)

	// Add database identity attributes for Service Graph
	opts = append(opts, trace.WithAttributes(semconv.DBSystemKey.String(t.dbSystem)))
	if t.dbNamespace != "" {
		opts = append(opts, trace.WithAttributes(
			semconv.DBNamespace(t.dbNamespace),
			// Versi lama fallback
			attribute.String("db.name", t.dbNamespace),
			//  Kunci utama agar 'server' di Service Graph terisi nama DB Anda
			attribute.String("peer.service", t.dbNamespace),
		))
	} else {
		// Jika namespace kosong, set default nama peer.service agar tidak kosong di Grafana
		opts = append(opts, trace.WithAttributes(
			attribute.String("peer.service", t.dbSystem),
		))
	}

	if t.serverAddress != "" {
		opts = append(opts, trace.WithAttributes(
			semconv.ServerAddress(t.serverAddress),
			// network peer address format lama yang sering dibaca oleh Alloy/Tempo v1.x
			attribute.String("net.peer.name", t.serverAddress),
		))
	}

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
