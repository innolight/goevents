package kafka

import (
	"context"
	"github.com/innolight/goevents"
	"github.com/segmentio/kafka-go"
	"time"
)

func NewProvider(addresses ...string) goevents.QueueProvider {
	return provider{addresses: addresses}
}

type provider struct {
	addresses []string
}

func (p provider) Get(ctx context.Context, kafkaTopic string) (goevents.Queue, error) {
	return &topic{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(p.addresses...),
			Topic:        kafkaTopic,
			RequiredAcks: kafka.RequireAll,
		},
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: p.addresses,
			Topic:   kafkaTopic,
			GroupID: "my-group",
			MaxWait: time.Second,
		}),
	}, nil
}
