package natsx

import (
	"time"

	"github.com/nats-io/nats.go"
)

type sub struct {
	s *nats.Subscription
}

func wrapSub(s *nats.Subscription) Subscription {
	return &sub{
		s: s,
	}
}

func (s *sub) NextMsg(timeout time.Duration) (*Msg, error) {
	return s.s.NextMsg(timeout)
}
func (s *sub) Pending() (int, int, error) {
	return s.s.Pending()
}
func (s *sub) IsValid() bool {
	return s.s.IsValid()
}
func (s *sub) AutoUnsubscribe(max int) error {
	return s.s.AutoUnsubscribe(max)
}
func (s *sub) Unsubscribe() error {
	return s.s.Unsubscribe()
}
func (s *sub) Drain() error {
	return s.s.Drain()
}
