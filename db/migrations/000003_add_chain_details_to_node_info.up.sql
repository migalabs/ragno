-- Add new columns on node_info about node's chain details
ALTER TABLE node_info ADD COLUMN fork_id TEXT;
ALTER TABLE node_info ADD COLUMN protocol_version INT;
ALTER TABLE node_info ADD COLUMN head_hash TEXT;
ALTER TABLE node_info ADD COLUMN network_id INT;
ALTER TABLE node_info ADD COLUMN total_difficulty BIGINT;
