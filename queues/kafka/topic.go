package kafka

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/innolight/goevents"
	"github.com/segmentio/kafka-go"
	"log"
)

type topic struct {
	writer *kafka.Writer
	reader *kafka.Reader
}

func (t *topic) Send(ctx context.Context, event goevents.Event) error {
	bytes, err := event.Body()
	if err != nil {
		return err
	}
	err = t.writer.WriteMessages(ctx, kafka.Message{
		Key: []byte(uuid.NewString()),
		// create an arbitrary message payload for the value
		Value: bytes,
	})
	if err != nil {
		return fmt.Errorf("writer.WriteMessages: %w", err)
	}

	return nil
}

func (t *topic) Receive(ctx context.Context) ([]goevents.EventEnvelop, error) {
	msg, err := t.reader.FetchMessage(ctx)
	if err != nil {
		return nil, err
	}

	resultChannel := make(chan error)
	go func() {
		err := <-resultChannel
		close(resultChannel)
		if err != nil {
			log.Println("[ERROR] handling message")
			// TODO: how to redeliver the message?
			return
		}
		// TODO: add logging and retry for acknowledge messages
		if err = t.reader.CommitMessages(context.TODO(), msg); err != nil {
			log.Println("[ERROR] committing message")
			// TODO: how to redeliver the message
			return
		}

		log.Printf("Acknowledged processed message: %s\n", string(msg.Value))
	}()

	return []goevents.EventEnvelop{
		{
			Result: resultChannel,
			Event:  goevents.StringEvent(ctx, string(msg.Value)),
		},
	}, nil
}
