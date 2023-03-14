# Streams Driver for SQL

The **stream driver** for `SQL` offers a `Writer` implementation to be used by systems implementing the
_**transactional outbox**_ messaging pattern.

Moreover, the `Message Egress Proxy` (_aka. log trailing_) component could be used along this driver to publish the 
messages to the message broker / stream.

Finally, depending on the underlying database engine, the `Message Egress Proxy` component could be combined with
a Change-Data-Capture (CDC) listener sidecar daemon.

This sidecar daemon could listen directly to CDC logs issued by the database engine 
(_e.g. WAL in PSQL, binlog in MySQL_). 

Another way avoid constant database polling is by implementing an agnostic sidecar which is deployed along the main 
container; it would accept connections through OSI Level 4 sockets (e.g. TCP, UDP, UNIX). 
The application container would send signals (heartbeats) to the sidecar daemon to trigger a data polling process.
Lastly, the daemon would use `Message Egress Proxy` to forward message batches to actual data-in-motion infrastructure.

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
    message_count INTEGER DEFAULT 0,
    raw_data BYTEA NOT NULL
);
```
