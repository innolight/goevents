## Example: Wishlist and Recommendation

In this example system, we have two applications:

+ The `Wishlist` application implements an HTTP endpoint, which sends added-wishlist-item events
  to `sns-demo-added-wishlist-items` SNS topic
+ SQS queue `sqs-demo-added-wishlist-items` subscribes to the SNS topic
+ The `Recommendation` application receives added-wishlist-item events from the `sqs-demo-added-wishlist-items` queue,
  computes recommendations, and sends recommendations to SQS queue `sqs-demo-recommendations`

## How to run

Start `localstack` container to mock SNS and SQS locally by executing

```bash
docker-compose up
```

Create necessary SQS queues and SNS topics, by running:

```
./example/sqssns/init.sh
```

Run wishlist application:

```bash
go run example/snssqs/wishlist/main.go
```

Run recommendation application:

```bash
go run example/snssqs/recommendation/main.go
```

Invoke the wishlist endpoint with curl:

```bash
curl http://localhost:8080/wishlists/123/items -XPOST
```

After this you can see the logs for the wishlist and recommendation services.
