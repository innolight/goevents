package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/google/uuid"
	"github.com/innolight/goevents"
	"github.com/innolight/goevents/middlewares"
	"github.com/innolight/goevents/queues/sqs"
	"time"
)

const (
	addedWishlistItemsTopic = "sqs-demo-added-wishlist-items"
	recommendationsTopic    = "sqs-demo-recommendations"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()
	transport := goevents.NewTransport(
		sqs.NewQueueProvider(sqs.WithAwsConfig(&aws.Config{Endpoint: aws.String("http://localhost:4566"), Region: aws.String("eu-central-1")})),
		goevents.WithMiddlewares(
			middlewares.LoggingMiddleware(),
		),
	)
	defer func() {
		ctx, _ := context.WithTimeout(context.Background(), time.Second*30)
		transport.Shutdown(ctx)
	}()

	addedWishlistItems, err := transport.Receive(addedWishlistItemsTopic)
	handleErr(err)
	recommendations, err := transport.Send(recommendationsTopic)
	handleErr(err)

	fmt.Println("Application started...")
	Recommend(ctx, addedWishlistItems, recommendations)
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func Recommend(ctx context.Context, wishlistItems <-chan goevents.EventEnvelop, recos chan<- goevents.EventEnvelop) {
	for {
		select {
		case wishlistItem := <-wishlistItems:
			sendResult := make(chan error, 1)
			recos <- goevents.EventEnvelop{
				Result: sendResult,
				Event:  goevents.JsonEvent(wishlistItem.Context(), getRecommendation(wishlistItem.Event)),
			}
			wishlistItem.Result <- <-sendResult
		case <-ctx.Done():
			return
		}
	}
}

func getRecommendation(event goevents.Event) recommendationEvent {
	var e wishlistItemAddedEvent
	if err := goevents.BindJson(event, &e); err != nil {
		panic(err)
	}
	return recommendationEvent{
		User:             e.UserID,
		RecommendedItems: []string{uuid.NewString()},
	}
}

type recommendationEvent struct {
	User             string   `json:"user"`
	RecommendedItems []string `json:"recommended_items"`
}

type wishlistItemAddedEvent struct {
	UserID string `json:"user_id"`
	ItemID string `json:"item_id"`
}
