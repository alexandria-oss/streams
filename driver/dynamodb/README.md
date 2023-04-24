# Streams Driver for Amazon DynamoDB

The **stream driver** for `Amazon DynamoDB` which offers a deduplication storage implementation to ensure idempotency
for message processing.

It is planned to offer in a near future a `Writer` implementation to be used by systems implementing the
_**transactional outbox**_ messaging pattern.

Moreover, the `Message Egress Proxy` (_aka. log trailing_) component could be used along this driver to 
publish the messages to the message broker / stream.

Furthermore, `Amazon DynamoDB` has a **_Change-Data-Capture stream feature_** ready to write changes into multiple 
services such as _Lambda and Kinesis_. Thus, this feature could be combined along the `Message Egress Proxy` component
in order to stream messages into desired infrastructure non-supported by `Amazon DynamoDB Stream` feature.

## Deduplication Storage Requirements

In order for this driver to work, the database MUST have a deduplication table with the following schema.

```json
{
  "TableName": "deduplication-table",
  "KeySchema": [
    {
      "KeyType": "HASH",
      "AttributeName": "worker_id"
    },
    {
      "KeyType": "RANGE",
      "AttributeName": "message_id"
    }
  ],
  "AttributeDefinitions": [
    {
      "AttributeName": "worker_id",
      "AttributeType": "S"
    },
    {
      "AttributeName": "message_id",
      "AttributeType": "S"
    }
  ],
  "BillingMode": "PAY_PER_REQUEST"
}
```

## Transactional Outbox Requirements

In order for this driver to work, the database MUST have an outbox table with the following schema.

```json
{
  "TableName": "deduplication-table",
  "KeySchema": [
    {
      "KeyType": "HASH",
      "AttributeName": "worker_id"
    },
    {
      "KeyType": "RANGE",
      "AttributeName": "message_id"
    }
  ],
  "AttributeDefinitions": [
    {
      "AttributeName": "worker_id",
      "AttributeType": "S"
    },
    {
      "AttributeName": "message_id",
      "AttributeType": "S"
    }
  ],
  "BillingMode": "PAY_PER_REQUEST"
}
```
