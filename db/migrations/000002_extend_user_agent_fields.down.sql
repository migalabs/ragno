-- IMPORTANT: order matters (first drop, then rename)

-- drop extra client user-agent details
ALTER TABLE node_info DROP COLUMN IF EXISTS client_name;
ALTER TABLE node_info DROP COLUMN IF EXISTS client_raw_version;
ALTER TABLE node_info DROP COLUMN IF EXISTS client_clean_version;
ALTER TABLE node_info DROP COLUMN IF EXISTS client_os;
ALTER TABLE node_info DROP COLUMN IF EXISTS client_arch;
ALTER TABLE node_info DROP COLUMN IF EXISTS client_language;

-- rollback column rename
ALTER TABLE node_info RENAME COLUMN client_name TO raw_user_agent;