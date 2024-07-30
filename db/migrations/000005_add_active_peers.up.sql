-- Create active_peers table for snapshots
CREATE TABLE IF NOT EXISTS active_peers (
    id SERIAL,
    timestamp TIMESTAMP,
    peers BIGINT[],

    PRIMARY KEY(timestamp)
);
