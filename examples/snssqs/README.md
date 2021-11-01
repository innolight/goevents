## Example: Wishlist and Recommendation

In this example system, we have two applications:

+ The `Wishlist` application implements an HTTP endpoint, which sends added-wishlist-item events
  to `sns-demo-added-wishlist-items` SNS topic
+ SQS queue `sqs-demo-added-wishlist-items` subscribes to the SNS topic
+ The `Recommendation` application receives added-wishlist-item events from the `sqs-demo-added-wishlist-items` queue,
  computes recommendations, and sends recommendations to SQS queue `sqs-demo-recommendations`

## How to run

Start `localstack` container to mock SNS and SQS locally by running docker-compose:

```bash
docker-compose -f examples/snssqs/docker-compose.yaml up
```

Create necessary SQS queues and SNS topics, by running:

```
./examples/snssqs/init_snssqs.sh
```

Run wishlist application:

```bash
go run examples/snssqs/wishlist/main.go
```

Run recommendation application:

```bash
go run examples/snssqs/recommendation/main.go
```

Invoke the wishlist endpoint with curl:

```bash
curl http://localhost:8080/wishlists/123/items -XPOST
```

After this you can see the logs for the wishlist and recommendation services.

Or run a load test against the endpoint, .e.g by using vegeta:
```bash
echo "POST http://localhost:8080/wishlists/123/items" | vegeta attack -duration=300s -rate 100/1s | vegeta encode | \
    jaggr @count=rps \
          hist\[100,200,300,400,500\]:code \
          p25,p50,p95:latency | \
    jplot rps+code.hist.100+code.hist.200+code.hist.300+code.hist.400+code.hist.500 \
          latency.p95+latency.p50+latency.p25
```