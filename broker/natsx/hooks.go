package natsx

import (
	"context"
	"time"
)

// Hook captures lifecycle events around NATS operations for logging/tracing/metrics.
// Implementors should be lightweight and non-blocking.
// All methods should be safe for concurrent use.
type Hook interface {
	// --- Publish / Request ---
	BeforePublish(ctx context.Context, subject string, msg *Msg) context.Context
	AfterPublish(ctx context.Context, subject string, msg *Msg, err error, took time.Duration)

	BeforeRequest(ctx context.Context, subject string, msg *Msg) context.Context
	AfterRequest(ctx context.Context, subject string, req *Msg, resp *Msg, err error, took time.Duration)

	// --- Subscribe lifecycle ---
	AfterSubscribe(ctx context.Context, subject, queue string, err error)
	AfterSubscribeSync(ctx context.Context, subject, queue string, err error)
	AfterChanSubscribe(ctx context.Context, subject, queue string, err error)

	// --- Message handling (async callbacks) ---
	BeforeHandle(ctx context.Context, subject string, msg *Msg) context.Context
	AfterHandle(ctx context.Context, subject string, msg *Msg, err error, took time.Duration)
}
