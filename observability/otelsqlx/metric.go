package otelsqlx

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

// createMetrics initializes all synchronous metrics tracked by Tracer.
// Any errors encountered upon metric creation will be sent to the globally assigned OpenTelemetry ErrorHandler.
func (t *Tracer) createMetrics() {
	var err error

	t.operationDuration, err = t.meter.Int64Histogram(
		semconv.DBClientOperationDurationName,
		metric.WithDescription(semconv.DBClientOperationDurationDescription),
		metric.WithUnit("ms"),
	)
	if err != nil {
		otel.Handle(err)
	}

	t.operationErrors, err = t.meter.Int64Counter(
		string(DBClientOperationErrorsKey),
		metric.WithDescription("The count of database client operation errors"),
	)
	if err != nil {
		otel.Handle(err)
	}
}

// incrementOperationErrorCount will increment the operation error count metric for any provided error
// that is non-nil and not sql.ErrNoRows. Otherwise, incrementOperationErrorCount becomes a no-op.
func (t *Tracer) incrementOperationErrorCount(ctx context.Context, err error, operation string) {
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.operationErrors.Add(ctx, 1, metric.WithAttributeSet(
			attribute.NewSet(append(t.meterAttrs, OperationTypeKey.String(operation))...),
		))
	}
}

// recordOperationDuration will compute and record the time since the start of an operation.
func (t *Tracer) recordOperationDuration(ctx context.Context, operation string, startTime time.Time) {
	t.operationDuration.Record(ctx, time.Since(startTime).Milliseconds(), metric.WithAttributeSet(
		attribute.NewSet(append(t.meterAttrs, OperationTypeKey.String(operation))...),
	))
}
