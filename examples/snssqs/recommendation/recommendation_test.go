package main

import (
	"context"
	"github.com/innolight/goevents"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRecommendation(t *testing.T) {
	// GIVEN
	wishlistItems := make(chan goevents.EventEnvelop)
	recommendations := make(chan goevents.EventEnvelop)
	go Recommend(context.Background(), wishlistItems, recommendations)

	// WHEN
	wishlistItems <- goevents.EventEnvelop{
		Result: make(chan error, 1),
		Event:  goevents.JsonEvent(context.Background(), wishlistItemAddedEvent{UserID: "user1"}),
	}

	// THEN
	select {
	case <-time.After(time.Millisecond * 100):
		t.Fail()
	case reco := <-recommendations:
		var result recommendationEvent
		assert.NoError(t, goevents.BindJson(reco, &result))
		assert.Equal(t, result.User, "user1")
	}
}
