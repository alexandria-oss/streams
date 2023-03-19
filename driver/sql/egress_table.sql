-- Using KSUID as primary key, hence the CHAR(27) type.
CREATE TABLE IF NOT EXISTS streams_egress(
    batch_id VARCHAR(27) PRIMARY KEY,
    message_count INTEGER DEFAULT 0,
    raw_data BYTEA NOT NULL
);
