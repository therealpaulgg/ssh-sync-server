package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
	"github.com/therealpaulgg/ssh-sync-server/test/pgx"
)

func TestSshKeyRepo_UpsertSshKeyWithChange(t *testing.T) {
	injector := do.New()
	db := pgx.NewMockDatabase()

	// Mock the query service
	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.SshKey], error) {
		return &query.QueryServiceImpl[models.SshKey]{
			DataAccessor: db,
		}, nil
	})

	// Mock the transaction service
	txMock := &query.TransactionMock{}
	do.Provide(injector, func(i *do.Injector) (query.TransactionService, error) {
		return txMock, nil
	})

	// Mock for change repository
	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.SshKeyChange], error) {
		return &query.QueryServiceImpl[models.SshKeyChange]{
			DataAccessor: db,
		}, nil
	})

	do.Provide(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshKeyChange], error) {
		return &query.QueryServiceTxImpl[models.SshKeyChange]{
			DataAccessor: db,
		}, nil
	})

	// Create repo under test
	repo := SshKeyRepo{
		Injector: injector,
	}

	// Test data
	userID := uuid.New()
	keyID := uuid.New()
	sshKey := &models.SshKey{
		UserID:   userID,
		Filename: "id_rsa.pub",
		Data:     []byte("ssh-rsa AAAAB3NzaC..."),
	}

	// Mock behavior for creating a new key
	db.MockQueryRow = func(sql string, args ...interface{}) []interface{} {
		if sql == "SELECT * FROM ssh_keys WHERE user_id = $1 AND filename = $2" {
			// No existing key found
			return nil
		}
		if sql == "INSERT INTO ssh_keys (user_id, filename, data) VALUES ($1, $2, $3) ON CONFLICT (user_id, filename) DO UPDATE SET data = $3 RETURNING *" {
			return []interface{}{
				keyID,
				userID,
				"id_rsa.pub",
				[]byte("ssh-rsa AAAAB3NzaC..."),
			}
		}
		if sql == "INSERT INTO ssh_key_changes (id, ssh_key_id, user_id, change_type, filename, previous_data, new_data, change_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *" {
			assert.Equal(t, keyID, args[1])
			assert.Equal(t, userID, args[2])
			assert.Equal(t, models.Created, args[3])
			assert.Equal(t, "id_rsa.pub", args[4])
			assert.Nil(t, args[5]) // No previous data for new key
			assert.Equal(t, []byte("ssh-rsa AAAAB3NzaC..."), args[6])
			
			return []interface{}{
				args[0], // ID
				keyID,
				userID,
				models.Created,
				"id_rsa.pub",
				nil,
				[]byte("ssh-rsa AAAAB3NzaC..."),
				args[7], // ChangeTime
			}
		}
		return nil
	}

	// Test Create
	result, err := repo.UpsertSshKeyWithChange(sshKey)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, keyID, result.ID)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, "id_rsa.pub", result.Filename)
	assert.Equal(t, []byte("ssh-rsa AAAAB3NzaC..."), result.Data)

	// Mock behavior for updating an existing key
	existingKey := &models.SshKey{
		ID:       keyID,
		UserID:   userID,
		Filename: "id_rsa.pub",
		Data:     []byte("ssh-rsa OLDKEY..."),
	}
	
	updatedKey := &models.SshKey{
		UserID:   userID,
		Filename: "id_rsa.pub",
		Data:     []byte("ssh-rsa NEWKEY..."),
	}

	db.MockQueryRow = func(sql string, args ...interface{}) []interface{} {
		if sql == "SELECT * FROM ssh_keys WHERE user_id = $1 AND filename = $2" {
			return []interface{}{
				keyID,
				userID,
				"id_rsa.pub",
				[]byte("ssh-rsa OLDKEY..."),
			}
		}
		if sql == "INSERT INTO ssh_keys (user_id, filename, data) VALUES ($1, $2, $3) ON CONFLICT (user_id, filename) DO UPDATE SET data = $3 RETURNING *" {
			return []interface{}{
				keyID,
				userID,
				"id_rsa.pub",
				[]byte("ssh-rsa NEWKEY..."),
			}
		}
		if sql == "INSERT INTO ssh_key_changes (id, ssh_key_id, user_id, change_type, filename, previous_data, new_data, change_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *" {
			assert.Equal(t, keyID, args[1])
			assert.Equal(t, userID, args[2])
			assert.Equal(t, models.Updated, args[3])
			assert.Equal(t, "id_rsa.pub", args[4])
			assert.Equal(t, []byte("ssh-rsa OLDKEY..."), args[5])
			assert.Equal(t, []byte("ssh-rsa NEWKEY..."), args[6])
			
			return []interface{}{
				args[0], // ID
				keyID,
				userID,
				models.Updated,
				"id_rsa.pub",
				[]byte("ssh-rsa OLDKEY..."),
				[]byte("ssh-rsa NEWKEY..."),
				args[7], // ChangeTime
			}
		}
		return nil
	}

	// Test Update
	result, err = repo.UpsertSshKeyWithChange(updatedKey)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, keyID, result.ID)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, "id_rsa.pub", result.Filename)
	assert.Equal(t, []byte("ssh-rsa NEWKEY..."), result.Data)
}
