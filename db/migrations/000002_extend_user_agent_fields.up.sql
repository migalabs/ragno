-- IMPORTANT: order matters (first rename, then add)

-- rename column for the raw-client-name / user-agent
ALTER TABLE node_info RENAME COLUMN client_name TO raw_user_agent;

-- add new columns with further details
ALTER TABLE node_info ADD COLUMN client_name TEXT;
ALTER TABLE node_info ADD COLUMN client_raw_version TEXT;
ALTER TABLE node_info ADD COLUMN client_clean_version TEXT;
ALTER TABLE node_info ADD COLUMN client_os TEXT;
ALTER TABLE node_info ADD COLUMN client_arch TEXT;
ALTER TABLE node_info ADD COLUMN client_language TEXT;
