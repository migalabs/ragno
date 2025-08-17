-- Create table to store every connection attempt
CREATE TABLE IF NOT EXISTS conn_attempts (
  id            SERIAL PRIMARY KEY,
  node_id       TEXT NOT NULL REFERENCES node_info(node_id),
  tried_at      TIMESTAMPTZ NOT NULL,
  error         TEXT,
  deprecated    BOOLEAN,
  latency       BIGINT
);
