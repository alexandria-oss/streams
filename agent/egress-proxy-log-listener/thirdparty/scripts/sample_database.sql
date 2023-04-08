-- Using KSUID as primary key, hence the CHAR(27) type.
CREATE TABLE IF NOT EXISTS streams_egress(
     batch_id VARCHAR(27) PRIMARY KEY,
     message_count INTEGER DEFAULT 0,
     raw_data BYTEA NOT NULL,
     insert_time TIMESTAMP DEFAULT NOW()
);

-- SELECT pg_create_logical_replication_slot('streams_egress_proxy', 'pgoutput');

CREATE USER streams_egress_proxy_replicator WITH REPLICATION PASSWORD 'foobar';

GRANT ALL ON SCHEMA public TO streams_egress_proxy_replicator;

GRANT SELECT ON public.streams_egress TO streams_egress_proxy_replicator;
GRANT DELETE ON public.streams_egress TO streams_egress_proxy_replicator;

CREATE PUBLICATION streams_egress_proxy FOR TABLE public.streams_egress ;
