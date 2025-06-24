package observability

import (
	"time"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type OptionParams struct {
	ServiceName  string
	Env          string
	OtlpEndpoint string
	OtlpUsername string
	OtlpPassword string
}

func NewObservabilityUsingKafka(params OptionParams) (trace.Tracer, func(), error) {
	closeFunc, err := NewOtel(params.ServiceName, params.OtlpEndpoint, params.OtlpUsername, params.OtlpPassword)
	if err != nil {
		return nil, nil, err
	}

	return otel.Tracer(params.ServiceName), func() {
		closeFunc()
	}, err
}

func NewLogWithKafkaHook(kafkaAddr []string, transportKafka *kafka.Transport, topic string, optionsParams OptionParams) func() {
	w := &kafka.Writer{
		Addr:            kafka.TCP(kafkaAddr...),
		Topic:           topic,
		Balancer:        &kafka.LeastBytes{},
		MaxAttempts:     5,
		WriteBackoffMin: time.Duration(100),
		WriteBackoffMax: time.Duration(1 * time.Second),

		BatchSize:    10,
		BatchBytes:   1048576,
		BatchTimeout: time.Duration(3 * time.Second),

		RequiredAcks: kafka.RequireOne,
		Transport:    transportKafka,
	}
	NewLog(&KafkaHook{
		writer:      w,
		topic:       topic,
		env:         optionsParams.Env,
		serviceName: optionsParams.ServiceName,
	}, optionsParams.Env, optionsParams.ServiceName)

	return func() {
		w.Close()
	}
}
