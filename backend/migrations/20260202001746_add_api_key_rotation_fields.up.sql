-- Migration up: add_api_key_rotation_fields
-- Add version control and rotation fields to user_api_keys table

-- Add parent_key_id for tracking key rotation chain
ALTER TABLE user_api_keys 
ADD COLUMN IF NOT EXISTS parent_key_id UUID REFERENCES user_api_keys(id) ON DELETE SET NULL;

-- Add version number for tracking key versions
ALTER TABLE user_api_keys 
ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;

-- Add rotation metadata
ALTER TABLE user_api_keys 
ADD COLUMN IF NOT EXISTS rotation_reason TEXT;

ALTER TABLE user_api_keys 
ADD COLUMN IF NOT EXISTS rotated_at TIMESTAMP WITH TIME ZONE;

-- Create index for faster queries on rotation chain
CREATE INDEX IF NOT EXISTS idx_user_api_keys_parent_key_id ON user_api_keys(parent_key_id);
CREATE INDEX IF NOT EXISTS idx_user_api_keys_version ON user_api_keys(user_id, version);

-- Update existing keys to have version 1
UPDATE user_api_keys SET version = 1 WHERE version IS NULL;