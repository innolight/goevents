#!/bin/bash
set -euo pipefail

aws --endpoint-url=http://localhost:4566 sns create-topic --name sns-demo-added-wishlist-items

aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name sqs-demo-added-wishlist-items
aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name sqs-demo-recommendations

aws --endpoint-url=http://localhost:4566 sns subscribe --topic-arn arn:aws:sns:eu-central-1:000000000000:sns-demo-added-wishlist-items --protocol sqs --notification-endpoint http://localhost:4566/000000000000/sqs-demo-added-wishlist-items
