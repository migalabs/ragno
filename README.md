# Ragno

A simple yet-powerful and updated crawler for Ethereum's EL.
It has some basics to crawl and connect to EL nodes.

# Installation
```
# Download the git repository recursively (to update submodules)
git clone --recurse-submodules git@github.com:migalabs/ragno.git && cd ragno

# Install
make dependencies && make install
```

# Usage
The tool supports the following options:
```
USAGE:
   ragno [commands] [options...]

COMMANDS:
   discv4   discover4 prints nodes in the discovery4 network
   run      run connects to nodes provided in csv file and save into postgresql database
   connect  connect and identify any given ENR
   help, h  Shows a list of commands or help for one command


GLOBAL OPTIONS:
   --help, -h  show help
```

# Environment variables
More specific parameters can be also be configured.
When running with Docker, they are set according to your `.env` file.
Otherwise, setting them with flags directly is necessary.

```
OPTIONS:

--log-level                (string)    Defines the log level of the logs. ("trace", "info", "debug")
--db-endpoint              (string)    Complete endpoint of the database where the recollected data will be saved.
--ip                       (int)       IP to assign to the host.
--port                     (int)       Port to assign to the host.
--metrics-ip               (string)    IP where Prometheus metrics will be hosted.
--metrics-port             (int)       Port where Prometheus metrics will be exposed.
--metrics-endpoint         (string)    Endpoint where Prometheus metrics will be hosted.
--dialers, -cd             (int)       Amount of concurrent dialers for node connections.
--persisters, -cs          (int)       Amount of database writers.
--conn-timeout, -ct        (string)    Time to wait until a connection attempt is considered timed-out.
--snapshot-interval, -si   (string)    How often to insert into the `active_peers` table (snapshots of active nodes).
--ip-api-url, -ipapi       (string)    Full template URL to the API used for retrieving detailed IP information(`ip-api.com`).
--deprecation-time, -dt    (string)    Time limit for reconnecting to nodes before labelling them as deprecated.
```

# Docker
#### Build and run the database alongside ragno:
These containers are configured with a `.env` file. See `.env.example` for examples on the parameters.

`docker compose up db ragno`

#### Build and run Prometheus metrics database (exposed at `:9090`):
This container is configured with a `./prometheus/prometheus.yml` file. View `./prometheus/prometheus.yml.example` for a template.

`docker compose up prometheus`

If any container fails to start, make sure you granted permissions to the data folder with:
`sudo chmod 777 ./app-data/*_db`

# Prometheus
[Prometheus](https://prometheus.io/docs/introduction/overview/) is what is used to gather metrics periodically from the recollected data. By default, they can be viewed at `:9070/metrics`.

#### Current available metrics:
- Client Distribution
- Client Version Distribution
- Geographical Distribution
- Node Distribution
- Deprecated Nodes
- OS Distribution
- Architechture Distribution
- Hosting Type Distribution
- RTT/Latency Distribution
- IPs Distribution

# Migrate
To move between database versions, use [go migrate](https://github.com/golang-migrate/migrate/).

In case of any database conflict, you can still force a specific version:

`migrate -path / -database "postgresql://username:secretkey@localhost:5432/database_name?sslmode=disable" force <version>`

If specific upgrades or downgrades need to be done manually, one could do this with:

`migrate -path database/migration/ -database "postgresql://user:password@localhost:5432/database_name?sslmode=disable" -verbose up`


# Output/Database tables info

#### `node_info`
Information about Execution Layer nodes.

| column                      | description |
|-----------------------------|-------------|
| `node_id`                   | The node's ID (decoded from the node's record). It is the primary key of the table.
| `pubkey`                    | The node's secp256k1 public key.
| `ip`                        | The node's IPv4 address.
| `tcp`                       | The node's TCP port.
| `first_connected`           | Timestamp of when the first successful connection to the node was made.
| `last_connected`            | Timestamp of when the last successful connection to the node was made.
| `last_tried`                | Timestamp of the last connection attempt.
| `raw_user_agent`            | The node's full user agent. (format `client/version/os-architechture/language version`).
| `capabilities`              | The node's capabilities/supported protocols.
| `error`                     | Latest connection error.
| `deprecated`                | Nodes will be marked as deprecated when no connection attempt to it was successful after 48 hours, or if the node is not from mainnet (`network ID 1`).
| `client_name`               | The node's client name.
| `client_raw_version`        | The node's full client version (with build info).
| `client_clean_version`      | The node's client version.
| `client_os`                 | Operating system of the node.
| `client_arch`               | Computer architecture of the node.
| `client_language`           | Language the client of the node is written in.
| `fork_id`                   | Fork ID the node follows.
| `protocol_version`          | Ethereum protocol version the node follows.
| `head_hash`                 | Hash of the latest block (head) the node sees.
| `network_id`                | Network ID of the network where the node resides.
| `total_difficulty`          | The total difficulty/cumulative measure of work up to the node.
| `latency`                   | Time in milliseconds between the latest successful connection attempt and the connection itself.

#### `active_peers`
Contains periodical snapshots of active nodes.
| column                      | description |
|-----------------------------|-------------|
| `timestamp`                 | Timestamp of when the snapshot was taken.
| `peers`                     | Array of the nodes active at that moment. The nodes are indexes referencing `node_info`'s `id` column.

#### `enrs`
Contains the response from the Discovery process. This information is used for connection attempts.
| column                      | description |
|-----------------------------|-------------|
| `node_id`                   | The node's ID (decoded from the node's record). It is the primary key of the table.
| `origin`                    | The discovery type the node uses (e.g. discv4).
| `first_seen`                | Timestamp of the first time the node was discovered.
| `last_seen`                 | Timestamp of the last time the node was seen.
| `ip`                        | The node's IPv4 address.
| `tcp`                       | The node's TCP port.
| `udp`                       | The node's UDP port.
| `seq`                       | The node's record sequence number.
| `pubkey`                    | The node's secp256k1 public key.
| `record`                    | The node's record.

#### `ip_info`
Contains more detailed information about the node's IP. The data gathered to populate this table is from `ip-api.com`.
| column                      | description |
|-----------------------------|-------------|
| `ip`                        | An IPv4 address. It is the primary key of the table.
| `expiration_time`           | Timestamp of the IP's expiration time.
| `continent`                 | The IP's continent name.
| `continent_code`            | The IP's continent ISO 3166-1 alpha-2 code.
| `country`                   | The IP's country name.
| `country_code`              | The IP's country ISO 3166-1 alpha-2 code.
| `region_name`               | The IP's region name.
| `region`                    | The IP's region ISO 3166-2 code.
| `city`                      | The IP's city name.
| `zip`                       | The IP's location zip code.
| `lat`                       | The IP's location's latitude in degrees.
| `lon`                       | The IP's location's longitude in degrees.
| `isp`                       | The Internet Service Provider of the IP.
| `org`                       | The organization that manages the IP address.
| `as_raw`                    | The full AS (Autonomous System) name the IP is part of.
| `asname`                    | The AS (Autonomous System) name the IP is part of.
| `mobile`                    | If the IP is associated with a mobile network or not.
| `proxy`                     | If the IP is associated with a proxy server.
| `hosting`                   | If the IP is associated with a hosting provider.

# Maintainer
@cortze

# Notes
Be careful with the input and output csv file names, since the `discv4` command will take `output.csv` by default.

Also, please note that the tool is currently in a developing stage. Any bugs report and/or suggestions are very welcome.
