package middlewares

import (
	"context"
	"errors"
	"github.com/cenkalti/backoff"
	"github.com/innolight/goevents"
	"log"
)

func RetryMiddleware(backoffPolicy func() backoff.BackOff) goevents.MiddleWare {
	return func(queue goevents.Topic) goevents.Topic {
		return retryWrapper{underlying: queue, backoffPolicy: backoffPolicy}
	}
}

type retryWrapper struct {
	underlying    goevents.Topic
	backoffPolicy func() backoff.BackOff
}

func (l retryWrapper) Send(ctx context.Context, event goevents.Event) error {
	count := 0
	return backoff.Retry(func() error {
		err := l.underlying.Send(ctx, event)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return backoff.Permanent(err)
		}
		if err != nil {
			log.Printf("[RETRY attempt-%d] Publish event [%v] failed with err: %s\n", count, event, err.Error())
			count++
		}
		return err
	}, l.backoffPolicy())
}

func (l retryWrapper) Receive(ctx context.Context) ([]goevents.EventEnvelop, error) {
	events, err := l.underlying.Receive(ctx)
	for _, e := range events {
		log.Printf("handling event: %v\n", e)
	}
	return events, err
}
