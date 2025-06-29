package libkafka

import (
	"slices"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/propagation"
)

// textMapCarrier is an implementation of OpenTelemetry's TextMapCarrier interface
// for injecting and extracting trace context from Kafka message headers.
type textMapCarrier struct {
	msg *kafka.Message
}

// Ensure that textMapCarrier implements the propagation.TextMapCarrier interface.
var _ propagation.TextMapCarrier = (*textMapCarrier)(nil)

// NewMsgCarrier returns a new textMapCarrier that wraps the given Kafka message.
//
// This is typically used for propagating OpenTelemetry trace context through
// Kafka message headers.
func NewMsgCarrier(msg *kafka.Message) *textMapCarrier {
	return &textMapCarrier{msg}
}

// Get retrieves the value associated with the given key from the Kafka message headers.
// If the key does not exist, an empty string is returned.
func (c *textMapCarrier) Get(key string) string {
	for _, h := range c.msg.Headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

// Set sets or replaces the value for the given key in the Kafka message headers.
// If the key already exists, it will be removed and replaced with the new value.
func (c *textMapCarrier) Set(key, value string) {
	// Ensure the uniqueness of the key by deleting existing headers with the same key.
	for i := len(c.msg.Headers) - 1; i >= 0; i-- {
		if c.msg.Headers[i].Key == key {
			c.msg.Headers = slices.Delete(c.msg.Headers, i, i+1)
		}
	}
	c.msg.Headers = append(c.msg.Headers, kafka.Header{
		Key:   key,
		Value: []byte(value),
	})
}

// Keys returns a list of all header keys currently stored in the Kafka message.
func (c *textMapCarrier) Keys() []string {
	out := make([]string, len(c.msg.Headers))
	for i, h := range c.msg.Headers {
		out[i] = h.Key
	}
	return out
}
