package main

import (
	"context"
	"fmt"
	"github.com/innolight/goevents"
	"github.com/innolight/goevents/middlewares"
	"github.com/innolight/goevents/queues/kafka"
	"math/rand"
	"time"
)

func main() {
	transport := goevents.NewTransport(
		kafka.NewProvider("localhost:29092"),
		goevents.WithMiddlewares(
			middlewares.LoggingMiddleware()),
	)
	topic := "topic-a"
	sendChan, err := transport.Send(topic)
	if err != nil {
		panic(err)
	}

	receiveChan, err := transport.Receive(topic)
	if err != nil {
		panic(err)
	}

	go func() {
		start := time.Now()
		for {
			sendChan <- goevents.EventEnvelop{
				Result: nil, // don't care about result
				Event:  goevents.StringEvent(context.Background(), fmt.Sprintf("(ping %d)", time.Now().Unix())),
			}
			time.Sleep(time.Duration(rand.Int63n(int64(time.Second * 10))))
			if time.Now().Sub(start) > time.Second*10 {
				break
			}
		}
	}()

	for message := range receiveChan {
		message.Result <- nil
	}
}
