#!/bin/bash
# Runs automatically inside the LocalStack container once it is ready.
# Creates the single-table-design table with a generic GSI1.
set -euo pipefail

TABLE="${DYNAMODB_TABLE:-boilerplate}"

echo "Creating DynamoDB table: ${TABLE}"

awslocal dynamodb create-table \
  --table-name "${TABLE}" \
  --attribute-definitions \
    AttributeName=PK,AttributeType=S \
    AttributeName=SK,AttributeType=S \
    AttributeName=GSI1PK,AttributeType=S \
    AttributeName=GSI1SK,AttributeType=S \
  --key-schema \
    AttributeName=PK,KeyType=HASH \
    AttributeName=SK,KeyType=RANGE \
  --global-secondary-indexes \
    '[{
        "IndexName": "GSI1",
        "KeySchema": [
          {"AttributeName": "GSI1PK", "KeyType": "HASH"},
          {"AttributeName": "GSI1SK", "KeyType": "RANGE"}
        ],
        "Projection": {"ProjectionType": "ALL"}
    }]' \
  --billing-mode PAY_PER_REQUEST

echo "Table ${TABLE} created."
