CREATE TABLE IF NOT EXISTS payments(
    payment_id VARCHAR(128) PRIMARY KEY,
    user_id VARCHAR(128) NOT NULL,
    amount MONEY DEFAULT 0.0
);

CREATE TABLE IF NOT EXISTS user_payments_stats(
    user_id VARCHAR(128) PRIMARY KEY,
    amount MONEY DEFAULT 0.0
);

CREATE TABLE IF NOT EXISTS streams_egress(
    batch_id CHAR(128) PRIMARY KEY,
    in_flight BOOLEAN DEFAULT FALSE,
    raw_data BYTEA NOT NULL
);