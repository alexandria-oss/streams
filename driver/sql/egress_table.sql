-- Using KSUID as primary key, hence the CHAR(27) type.
CREATE TABLE IF NOT EXISTS streams_egress(
    batch_id CHAR(27) PRIMARY KEY,
    in_flight BOOLEAN DEFAULT FALSE,
    raw_data BYTEA NOT NULL
);