package goevents

import "context"

type QueueProvider interface {
	// Get return the Queue identified by name to send and receive messages
	Get(ctx context.Context, name string) (Queue, error)
}

type Queue interface {
	// Send delivers message to the queue
	Send(ctx context.Context, event Event) error

	// Receive returns messages from the queue
	Receive(ctx context.Context) ([]EventEnvelop, error)
}

type MiddleWare func(queue Queue) Queue
