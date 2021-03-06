package middlewares

import (
	"context"
	"errors"
	"github.com/cenkalti/backoff"
	"github.com/innolight/goevents"
	"log"
)

func RetryMiddleware(backoffPolicy func() backoff.BackOff) goevents.MiddleWare {
	return func(queue goevents.Queue) goevents.Queue {
		return retryWrapper{underlying: queue, backoffPolicy: backoffPolicy}
	}
}

type retryWrapper struct {
	underlying    goevents.Queue
	backoffPolicy func() backoff.BackOff
}

func (l retryWrapper) Send(ctx context.Context, event goevents.Event) error {
	count := 0
	return backoff.Retry(func() error {
		err := l.underlying.Send(ctx, event)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return backoff.Permanent(err)
			}

			log.Printf("[RETRY attempt-%d] Publish event [%v] failed with err: %s\n", count, event, err.Error())
			count++
		}
		return err
	}, l.backoffPolicy())
}

func (l retryWrapper) Receive(ctx context.Context) ([]goevents.EventEnvelop, error) {
	return l.underlying.Receive(ctx)
}
