#!/bin/bash
set -euo pipefail

aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name sqs-demo-added-wishlist-items
aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name sqs-demo-recommendations
