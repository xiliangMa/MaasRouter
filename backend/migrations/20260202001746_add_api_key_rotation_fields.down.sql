-- Migration down: remove_api_key_rotation_fields
-- Remove version control and rotation fields from user_api_keys table

-- Drop indexes first
DROP INDEX IF EXISTS idx_user_api_keys_parent_key_id;
DROP INDEX IF EXISTS idx_user_api_keys_version;

-- Remove columns
ALTER TABLE user_api_keys 
DROP COLUMN IF EXISTS parent_key_id;

ALTER TABLE user_api_keys 
DROP COLUMN IF EXISTS version;

ALTER TABLE user_api_keys 
DROP COLUMN IF EXISTS rotation_reason;

ALTER TABLE user_api_keys 
DROP COLUMN IF EXISTS rotated_at;
