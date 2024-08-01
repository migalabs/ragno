-- Create ips table
CREATE TABLE IF NOT EXISTS ips (
    id SERIAL,
    ip TEXT NOT NULL,
    expiration_time TIMESTAMP NOT NULL,
    continent TEXT NOT NULL,
    continent_code TEXT NOT NULL,
    country TEXT NOT NULL,
    country_code TEXT NOT NULL,
    region TEXT NOT NULL,
    region_name TEXT NOT NULL,
    city TEXT NOT NULL,
    zip TEXT NOT NULL,
    lat REAL NOT NULL,
    lon REAL NOT NULL,
    isp TEXT NOT NULL,
    org TEXT NOT NULL,
    as_raw TEXT NOT NULL,
    asname TEXT NOT NULL,
    mobile BOOL NOT NULL,
    proxy BOOL NOT NULL,
    hosting BOOL NOT NULL,

    PRIMARY KEY (ip)
);