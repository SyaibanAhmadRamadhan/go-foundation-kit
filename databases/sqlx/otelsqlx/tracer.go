package otelsqlx

import (
	"runtime/debug"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName          = "github.com/SyaibanAhmadRamadhan/go-foundation-kit/databases/sqlx/otelsqlx"
	meterName           = "github.com/SyaibanAhmadRamadhan/go-foundation-kit/databases/sqlx/otelsqlx"
	startTimeCtxKey     = "otelpgxStartTime"
	sqlOperationUnknown = "UNKNOWN"
)

const (
	// RowsAffectedKey represents the number of rows affected.
	RowsAffectedKey = attribute.Key("sqlx.rows_affected")
	// QueryParametersKey represents the query parameters.
	QueryParametersKey = attribute.Key("sqlx.query.parameters")
	// PrepareStmtNameKey represents the prepared statement name.
	PrepareStmtNameKey = attribute.Key("sqlx.prepare_stmt.name")
	// SQLStateKey represents PostgreSQL error code,
	// see https://www.postgresql.org/docs/current/errcodes-appendix.html.
	SQLStateKey = attribute.Key("sqlx.sql_state")
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

	trimQuerySpanName    bool
	spanNameFunc         SpanNameFunc
	prefixQuerySpanName  bool
	logSQLStatement      bool
	logConnectionDetails bool
	includeParams        bool
}

type tracerConfig struct {
	tracerProvider trace.TracerProvider
	meterProvider  metric.MeterProvider

	tracerAttrs []attribute.KeyValue
	meterAttrs  []attribute.KeyValue

	trimQuerySpanName    bool
	spanNameFunc         SpanNameFunc
	prefixQuerySpanName  bool
	logSQLStatement      bool
	logConnectionDetails bool
	includeParams        bool
}

// NewTracer returns a new Tracer.
func NewTracer(opts ...Option) *Tracer {
	cfg := &tracerConfig{
		tracerProvider:       otel.GetTracerProvider(),
		meterProvider:        otel.GetMeterProvider(),
		tracerAttrs:          []attribute.KeyValue{},
		meterAttrs:           []attribute.KeyValue{},
		trimQuerySpanName:    false,
		spanNameFunc:         nil,
		prefixQuerySpanName:  true,
		logSQLStatement:      true,
		logConnectionDetails: true,
		includeParams:        false,
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
