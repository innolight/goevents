package middlewares

import (
	"context"
	"github.com/innolight/goevents"
	"log"
	"time"
)

func LoggingMiddleware() goevents.MiddleWare {
	return func(queue goevents.Topic) goevents.Topic {
		return loggingWrapper{queue}
	}
}

type loggingWrapper struct {
	underlying goevents.Topic
}

func (l loggingWrapper) Send(ctx context.Context, event goevents.Event) error {
	start := time.Now()
	err := l.underlying.Send(ctx, event)
	if err != nil {
		log.Printf("Failed to publish event [%v] after %s: %s\n", event, time.Since(start).Round(time.Millisecond), err.Error())
	} else {
		log.Printf("Published event: %v after %s\n", event, time.Since(start).Round(time.Millisecond))
	}
	return err
}

func (l loggingWrapper) Receive(ctx context.Context) ([]goevents.EventEnvelop, error) {
	events, err := l.underlying.Receive(ctx)
	for _, e := range events {
		log.Printf("Handling event: %v\n", e.Event)
	}
	return events, err
}
