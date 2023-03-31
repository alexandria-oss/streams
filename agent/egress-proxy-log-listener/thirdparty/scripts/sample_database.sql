-- Using KSUID as primary key, hence the CHAR(27) type.
CREATE TABLE IF NOT EXISTS streams_egress(
     batch_id VARCHAR(27) PRIMARY KEY,
     message_count INTEGER DEFAULT 0,
     raw_data BYTEA NOT NULL
);

SELECT pg_create_logical_replication_slot('streams_egress_proxy', 'pgoutput');

CREATE USER streams_egress_proxy_replicator WITH REPLICATION PASSWORD 'foobar';

GRANT ALL ON SCHEMA public TO streams_egress_proxy_replicator;

CREATE PUBLICATION streams_egress FOR TABLE public.streams_egress ;
