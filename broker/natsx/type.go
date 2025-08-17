package natsx

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
)

type (
	Msg        = nats.Msg
	MsgHandler = nats.MsgHandler
	SubOpt     = nats.SubOpt
	Status     = nats.Status
)

// Subscription membungkus nats.Subscription supaya gampang di-mock.
type Subscription interface {
	Unsubscribe() error
	Drain() error
	AutoUnsubscribe(max int) error
	// Sync ops:
	NextMsg(timeout time.Duration) (*Msg, error)
	// Monitoring:
	Pending() (int, int, error) // (msgs, bytes)
	IsValid() bool
}

// NATSClient adalah interface wrapper umum
type NATSClient interface {
	// --- Publish/Request ---
	Publish(ctx context.Context, subject string, data []byte) error
	PublishMsg(ctx context.Context, msg *Msg) error
	Request(ctx context.Context, subject string, data []byte, timeout time.Duration) (*Msg, error)
	RequestMsg(ctx context.Context, msg *Msg) (*Msg, error)

	// --- Subscribe (async callback) ---
	Subscribe(subject string, handler MsgHandler) (Subscription, error)
	QueueSubscribe(subject, queue string, handler MsgHandler) (Subscription, error)

	// --- Subscribe (sync: pull per-NextMsg) ---
	SubscribeSync(subject string) (Subscription, error)
	QueueSubscribeSync(subject, queue string) (Subscription, error)

	// --- Subscribe (channel-based) ---
	ChanSubscribe(subject string, ch chan *Msg) (Subscription, error)
	ChanQueueSubscribe(subject, queue string, ch chan *Msg) (Subscription, error)

	// --- Connection control/health ---
	Flush() error
	FlushTimeout(d time.Duration) error
	Drain() error
	Close()
	LastError() error
	Status() Status
	IsConnected() bool
	IsClosed() bool
	IsReconnecting() bool
}
