# Go channels for distributed systems

Benefits:
* Use Go channels transparently over a messaging technology of your choice
* Write idiomatic Go code instead of using vendor specific APIs
* Write simple and independent unit test (no need to mock interfaces or running the queue technology)

## Usage

The following code receives added-wishlist items from one queue and send recommendations to another queue.

```go
wishlistItems, _ := transport.Receive("added-wishlist-items")
recommendations, _ := transport.Send("recommendations")

for received := range wishlistItems {
        
        // Channel to subscribe to the result of sending event 
        sendResult := make(chan error, 1)

        recommendations <- goevents.EventEnvelop{
            Result: sendResult,
            Event:  goevents.JsonEvent(wishlistItem.Context(), getRecommendation(wishlistItem.Event)),
        }
        
        // Acknowledge the processing of the received event
        received.Result <- <-sendResult
    }
}
```

## Features

+ **Producers** can 
  + explicitly receive and handle the message's publishing result, enabling at-least-one delivery guarantee
  + Or fire-and-forget (removing the latency related to message publishing)
+ **Consumers** must acknowledge the processing of message to enable message redelivery (at-least-one processing guarantee).
+ Graceful shutdown hook to ensure all messages are flushed to the queue
+ **Middlewares** to customize the processing of events .e.g. error handling, logging, retry

## Supported queue technologies

+ [x] [Amazon SQS](https://aws.amazon.com/sqs/)
+ [x] Amazon SNS (only send of events)
+ [ ] Kafka
+ [ ] Rabbit MQ
+ [ ] Redis
+ [ ] Postgres database queue (support at-least-one delivery guarantee for producers with transaction)
