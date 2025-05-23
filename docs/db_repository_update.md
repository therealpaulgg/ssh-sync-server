# DB Repository Update Instructions

To fully implement the SSH key change history feature, the database schema needs to be updated in the [ssh-sync-db](https://github.com/therealpaulgg/ssh-sync-db) repository.

## Changes Required

Add the following SQL to the `init.sql` file in the ssh-sync-db repository:

```sql
-- Create enum type for change types
CREATE TYPE change_type AS ENUM ('created', 'updated', 'deleted');

-- Create table for tracking SSH key changes
CREATE TABLE IF NOT EXISTS ssh_key_changes (
    id UUID DEFAULT uuid_generate_v4() NOT NULL,
    ssh_key_id UUID NOT NULL,
    user_id UUID NOT NULL,
    change_type change_type NOT NULL,
    filename VARCHAR(255) NOT NULL,
    previous_data BYTEA,
    new_data BYTEA,
    change_time TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Add indexes for efficient lookups
CREATE INDEX IF NOT EXISTS idx_ssh_key_changes_ssh_key_id ON ssh_key_changes(ssh_key_id);
CREATE INDEX IF NOT EXISTS idx_ssh_key_changes_user_id ON ssh_key_changes(user_id);
CREATE INDEX IF NOT EXISTS idx_ssh_key_changes_change_time ON ssh_key_changes(change_time);

-- Add index for the most common query pattern: finding the latest changes per key for a user
CREATE INDEX IF NOT EXISTS idx_ssh_key_changes_user_time ON ssh_key_changes(user_id, change_time DESC);
```

## Implementation Notes

1. The schema matches what's defined in the `docs/sql/ssh_key_changes.sql` file in this repository
2. The schema uses the existing `uuid_generate_v4()` function that's already being used in the database
3. The table includes a foreign key reference to the users table with cascade deletion to ensure data integrity
4. Indexes have been added to optimize the most common query patterns

## Testing

After updating the init.sql file in the ssh-sync-db repository, you can test that the schema works correctly by:

1. Running `docker-compose up --build` in the ssh-sync-db repository
2. Connecting to the database with `psql -h localhost -p 5432 -U sshsync -d sshsync`
3. Verifying the table exists with `\d ssh_key_changes`