-- Schema for SSH Key Change tracking

-- Create enum type for change types
CREATE TYPE change_type AS ENUM ('created', 'updated', 'deleted');

-- Create table for tracking SSH key changes
CREATE TABLE IF NOT EXISTS ssh_key_changes (
    id UUID PRIMARY KEY,
    ssh_key_id UUID NOT NULL,
    user_id UUID NOT NULL,
    change_type change_type NOT NULL,
    filename VARCHAR(255) NOT NULL,
    previous_data BYTEA,
    new_data BYTEA,
    change_time TIMESTAMP WITH TIME ZONE NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Add indexes for efficient lookups
CREATE INDEX IF NOT EXISTS idx_ssh_key_changes_ssh_key_id ON ssh_key_changes(ssh_key_id);
CREATE INDEX IF NOT EXISTS idx_ssh_key_changes_user_id ON ssh_key_changes(user_id);
CREATE INDEX IF NOT EXISTS idx_ssh_key_changes_change_time ON ssh_key_changes(change_time);

-- Add index for the most common query pattern: finding the latest changes per key for a user
CREATE INDEX IF NOT EXISTS idx_ssh_key_changes_user_time ON ssh_key_changes(user_id, change_time DESC);