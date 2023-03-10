# Streams Driver for Amazon DynamoDB

The **stream driver** for `Amazon DynamoDB` offers a `Writer` implementation to be used by systems implementing the
_**transactional outbox**_ messaging pattern.

Moreover, the `Message Egress Proxy` (_aka. log trailing_) component could be used along this driver to 
publish the messages to the message broker / stream.

Furthermore, `Amazon DynamoDB` has a **_Change-Data-Capture stream feature_** ready to write changes into multiple 
services such as _Lambda and Kinesis_. Thus, this feature could be combined along the `Message Egress Proxy` component
in order to stream messages into desired infrastructure non-supported by `Amazon DynamoDB Stream` feature.

## Requirements

In order for this driver to work, the SQL database MUST have an outbox table with the following schema 
(_called streams_egress by default_).


```genericsql
-- Using KSUID as primary key, hence the CHAR(27) type.
--
-- NOTE: BYTEA type is a Postgres type used for binary array types.
CREATE TABLE IF NOT EXISTS streams_egress(
    batch_id CHAR(27) PRIMARY KEY,
    in_flight BOOLEAN DEFAULT FALSE,
    raw_data BYTEA NOT NULL
);
```
