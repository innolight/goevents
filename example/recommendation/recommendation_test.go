package main

import (
	"context"
	"github.com/innolight/goevents"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRecommend(t *testing.T) {
	wishlistItemsEvent := make(chan goevents.EventEnvelop)
	recommendationsEvents := make(chan goevents.EventEnvelop)

	go Recommend(context.Background(), wishlistItemsEvent, recommendationsEvents)

	ackItemProcessed := make(chan error, 1)
	wishlistItemsEvent <- goevents.EventEnvelop{
		Result: ackItemProcessed,
		Event:  goevents.Event{Body: []byte("user1 likes item1")},
	}

	recoEvent := <-recommendationsEvents
	assert.Equal(t, "user might like item2, item3", string(recoEvent.Body))

	recoEvent.Result <- nil // Ack event is sent
	assert.Nil(t, <-ackItemProcessed)
}
