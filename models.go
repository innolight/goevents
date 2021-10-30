package goevents

import "context"


type TopicFactory interface {
	Get(ctx context.Context, name string) (Topic, error)
}

type Topic interface {
	Send(ctx context.Context, event Event) error
	Receive(ctx context.Context) ([]EventEnvelop, error)
}

type MiddleWare func(queue Topic) Topic
