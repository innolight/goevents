package sns

import (
	"context"
	"encoding/json"
	"github.com/innolight/goevents"
)

var SNSNotificationParseMiddleware goevents.MiddleWare = func(queue goevents.Queue) goevents.Queue {
	return sqsMessageMapper{queue}
}

type sqsMessageMapper struct{ goevents.Queue }

func (s sqsMessageMapper) Send(ctx context.Context, event goevents.Event) error {
	return s.Queue.Send(ctx, event)
}

// Receive returns messages from the queue
func (s sqsMessageMapper) Receive(ctx context.Context) ([]goevents.EventEnvelop, error) {
	events, err := s.Queue.Receive(ctx)
	if err != nil {
		return nil, err
	}
	mappedEvents := make([]goevents.EventEnvelop, 0, len(events))
	for _, e := range events {
		bytes, err := e.Body()
		if err != nil {
			mappedEvents = append(mappedEvents, e)
			continue
		}
		var snsNotification struct {
			Message string `json:"Message"`
		}
		if err := json.Unmarshal(bytes, &snsNotification); err != nil {
			mappedEvents = append(mappedEvents, e)
			continue
		}
		mappedEvents = append(mappedEvents, goevents.EventEnvelop{
			Result: e.Result,
			Event:  goevents.StringEvent(e.Context(), snsNotification.Message),
		})

	}
	return mappedEvents, nil
}
