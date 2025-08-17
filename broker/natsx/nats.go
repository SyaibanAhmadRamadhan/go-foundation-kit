package natsx

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type client struct {
	nc *nats.Conn
}

type clientConfig struct {
}

var _ NATSClient = (*client)(nil)

func NewNats(opts ...Option) *client {
	nc, err := nats.Connect(nats.DefaultURL,
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			fmt.Println("Disconnected due to:", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			fmt.Println("Reconnected to", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			fmt.Println("Connection closed")
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			fmt.Printf("Async error in subscription %q: %v\n", sub.Subject, err)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Drain()
	return &client{
		nc: nc,
	}
}

// --- Publish/Request ---
func (c *client) Publish(ctx context.Context, subject string, data []byte) error {
	return c.nc.Publish(subject, data)
}
func (c *client) PublishMsg(ctx context.Context, msg *Msg) error {
	return c.nc.PublishMsg(msg)
}
func (c *client) Request(ctx context.Context, subject string, data []byte, t time.Duration) (*Msg, error) {
	return c.nc.Request(subject, data, t)
}
func (c *client) RequestMsg(ctx context.Context, msg *Msg) (*Msg, error) {
	return c.nc.RequestMsgWithContext(ctx, msg)
}

// --- Subscribe async ---
func (c *client) Subscribe(subject string, h MsgHandler) (Subscription, error) {
	sub, err := c.nc.Subscribe(subject, h)
	if err != nil {
		return nil, err
	}

	return wrapSub(sub), nil
}

func (c *client) QueueSubscribe(subject, queue string, h MsgHandler) (Subscription, error) {
	sub, err := c.nc.QueueSubscribe(subject, queue, h)
	if err != nil {
		return nil, err
	}
	return wrapSub(sub), nil
}

// --- Subscribe sync ---
func (c *client) SubscribeSync(subject string) (Subscription, error) {
	sub, err := c.nc.SubscribeSync(subject)
	if err != nil {
		return nil, err
	}

	return wrapSub(sub), nil
}

func (c *client) QueueSubscribeSync(subject, queue string) (Subscription, error) {
	sub, err := c.nc.QueueSubscribeSync(subject, queue)
	if err != nil {
		return nil, err
	}

	return wrapSub(sub), nil
}

// --- Subscribe channel-based ---
func (c *client) ChanSubscribe(subject string, ch chan *Msg) (Subscription, error) {
	sub, err := c.nc.ChanSubscribe(subject, ch)
	if err != nil {
		return nil, err
	}

	return wrapSub(sub), nil
}

func (c *client) ChanQueueSubscribe(subject, queue string, ch chan *Msg) (Subscription, error) {
	sub, err := c.nc.ChanQueueSubscribe(subject, queue, ch)
	if err != nil {
		return nil, err
	}

	return wrapSub(sub), nil
}

// --- Connection control ---
func (c *client) Flush() error {
	return c.nc.Flush()
}
func (c *client) FlushTimeout(d time.Duration) error {
	return c.nc.FlushTimeout(d)
}
func (c *client) Drain() error {
	return c.nc.Drain()
}
func (c *client) Close() {
	c.nc.Close()
}
func (c *client) LastError() error {
	return c.nc.LastError()
}
func (c *client) Status() Status {
	return c.nc.Status()
}
func (c *client) IsConnected() bool {
	return c.nc.IsConnected()
}
func (c *client) IsClosed() bool {
	return c.nc.IsClosed()
}
func (c *client) IsReconnecting() bool {
	return c.nc.IsReconnecting()
}
