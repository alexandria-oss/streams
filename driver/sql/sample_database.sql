CREATE TABLE IF NOT EXISTS payments(
    payment_id VARCHAR(128) PRIMARY KEY,
    user_id VARCHAR(128) NOT NULL,
    amount MONEY DEFAULT 0.0
);

-- Using KSUID as primary key, hence the CHAR(27) type.
CREATE TABLE IF NOT EXISTS streams_egress(
    batch_id VARCHAR(27) PRIMARY KEY,
    message_count INTEGER DEFAULT 0,
    raw_data BYTEA NOT NULL,
    insert_time TIMESTAMP DEFAULT NOW()
);
