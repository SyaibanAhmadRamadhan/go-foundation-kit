package libkafka

import (
	"context"
	"net"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

func NewTransportSasl(username, pass string) *kafka.Transport {
	mechanism := plain.Mechanism{
		Username: username,
		Password: pass,
	}

	return &kafka.Transport{
		SASL: mechanism,
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, addr)
		},
	}
}
