package main

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/cenkalti/backoff"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/innolight/goevents"
	"github.com/innolight/goevents/middlewares"
	"github.com/innolight/goevents/queues/sqs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const addedWishlistItemsTopic = "sqs-demo-added-wishlist-items"

func main() {
	awsConf := &aws.Config{Endpoint: aws.String("http://localhost:4566"), Region: aws.String("eu-central-1")}
	transport := goevents.NewTransport(sqs.NewTopicFactory(sqs.WithAwsConfig(awsConf)),
		goevents.WithMiddlewares(
			middlewares.LoggingMiddleware(),
			middlewares.RetryMiddleware(func() backoff.BackOff { return backoff.NewExponentialBackOff() }),
		),
		goevents.WithPublisherCount(400),
		goevents.WithBufferSize(1024*8),
	)
	topic, err := transport.Send(addedWishlistItemsTopic)
	if err != nil {
		panic(err)
	}

	router := gin.Default()

	router.POST("/wishlists/:wishlist-id/items", func(c *gin.Context) {
		// update database
		topic <- goevents.EventEnvelop{
			Event: goevents.JsonEvent(c.Request.Context(), getWishlistItemAddedEvent()),
		}
		c.JSON(http.StatusOK, "done")
	})

	runWithGracefulShutdown(&http.Server{
		Addr:    ":8080",
		Handler: router,
	})
	cleanUpCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	transport.Shutdown(cleanUpCtx)
}

func runWithGracefulShutdown(srv *http.Server) {
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}

type WishlistItemAddedEvent struct {
	UserID string `json:"user_id"`
	ItemID string `json:"item_id"`
}

func getWishlistItemAddedEvent() WishlistItemAddedEvent {
	return WishlistItemAddedEvent{
		UserID: uuid.NewString(),
		ItemID: uuid.NewString(),
	}
}
