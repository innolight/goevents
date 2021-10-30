package goevents

import (
	"context"
	"encoding/json"
	"fmt"
)

type EventEnvelop struct {
	// For sending message
	//   Result is an optional channel to receive the result of the message publishing, which is
	//     + nil if message was send successfully
	//     + error if the message couldn't be send
	//
	// For receiving message, Result channel is used to acknowledge the processing of message by receiver.
	//     + Receiver sends nil if message is processed successfully
	//     + Receiver sends error if the message cannot be processed
	Result chan<- error

	Event
}

type Event interface {
	// Context return the context from which the event is created
	// The returned context is used to propagate request-scoped values
	// Cancellation signals is not used
	Context() context.Context

	// Content of the event
	Body() ([]byte, error)
}

func BindJson(e Event, destination interface{}) error {
	bytes, err := e.Body()
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, destination)
}

func JsonEvent(ctx context.Context, payload interface{}) Event {
	return jsonEvent{
		ctx:  ctx,
		body: payload,
	}
}

type jsonEvent struct {
	ctx  context.Context
	body interface{}
}

func (j jsonEvent) Context() context.Context { return j.ctx }

func (j jsonEvent) Body() ([]byte, error) {
	return json.Marshal(j.body)
}

func (j jsonEvent) String() string {
	bytes, err := j.Body()
	if err != nil {
		return fmt.Sprintf("%+v", j.body)
	}
	return string(bytes)
}

func StringEvent(ctx context.Context, str string) Event {
	return stringEvent{ctx: ctx, body: str}
}

type stringEvent struct {
	ctx  context.Context
	body string
}

func (j stringEvent) Context() context.Context { return j.ctx }

func (j stringEvent) Body() ([]byte, error) {
	return []byte(j.body), nil
}

func (j stringEvent) String() string {
	return j.body
}
