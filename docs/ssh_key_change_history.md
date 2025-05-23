# SSH Key Change History

This feature adds support for tracking changes to SSH keys in the database, allowing for better conflict resolution during syncing.

## Database Schema

The feature introduces a new table `ssh_key_changes` that tracks all changes (creation, updates, and deletions) to SSH keys. 
See the SQL schema in `docs/sql/ssh_key_changes.sql`.

## API Usage

### Recording Changes

Changes are automatically recorded when using the new repository methods:

- `SshKeyRepo.CreateSshKeyWithChange`: Create a new SSH key and record it as a creation event
- `SshKeyRepo.UpsertSshKeyWithChange`: Create or update an SSH key and record the appropriate event
- `SshKeyRepo.UpsertSshKeyWithChangeTx`: Same as above but within a transaction
- `UserRepo.DeleteUserKeyTx`: Now records a deletion event before deleting the key

### Retrieving Change History

Use the `SshKeyChangeRepository` to access change history:

- `GetKeyChanges`: Retrieve the full change history for a specific SSH key
- `GetLatestKeyChangesForUser`: Get the most recent change for each of a user's SSH keys since a specified time

## Conflict Resolution

This change history enables better conflict resolution during syncing:

1. When a key is deleted on the server, clients can detect this change and delete it locally
2. When both server and client have made changes to the same key, the timestamps can be used to determine which change is more recent
3. In case of conflicts, the application can choose to keep the most recent change or prompt the user for resolution