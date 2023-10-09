-- Roll back node's chain details from node_info
ALTER TABLE node_info DROP COLUMN IF EXISTS fork_id;
ALTER TABLE node_info DROP COLUMN IF EXISTS protocol_version;
ALTER TABLE node_info DROP COLUMN IF EXISTS head_hash;
ALTER TABLE node_info DROP COLUMN IF EXISTS network_id;
ALTER TABLE node_info DROP COLUMN IF EXISTS total_difficulty;
