-- Create the table that will keep all the information gathered about the ENRs
CREATE TABLE IF NOT EXISTS enrs (
    id INT GENERATED ALWAYS AS IDENTITY,
    node_id TEXT PRIMARY KEY,
    origin TEXT NOT NULL,
    first_seen TIMESTAMP NOT NULL,
    last_seen TIMESTAMP NOT NULL,
    ip TEXT NOT NULL,
    tcp INT NOT NULL,
    udp INT NOT NULL,
    seq BIGINT NOT NULL,
    pubkey TEXT NOT NULL,
    record TEXT NOT NULL,
    score INT
);

-- Create the table that will keep the individual info of the EL nodes
CREATE TABLE IF NOT EXISTS node_info (
     id INT GENERATED ALWAYS AS IDENTITY,
     node_id TEXT PRIMARY KEY,
     pubkey TEXT NOT NULL,
     ip TEXT NOT NULL,
     tcp INT NOT NULL,
     first_connected TIMESTAMP,
     last_connected TIMESTAMP,
     last_tried TIMESTAMP,
     client_name TEXT,
     capabilities TEXT[],
     software_info INT,
     error TEXT,
     deprecated BOOL
);