package libkafka

import (
	"context"
	"net"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

// NewTransportSasl creates a new Kafka transport configured with SASL/PLAIN authentication.
//
// This is useful for connecting to a Kafka broker that requires SASL/PLAIN credentials,
// such as brokers hosted on cloud services or secured internal environments.
//
// Parameters:
//   - username: the SASL username
//   - pass: the SASL password
//
// Returns:
//   - *kafka.Transport: a transport instance with SASL and custom dialer support
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
