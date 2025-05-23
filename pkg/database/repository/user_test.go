package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
	"github.com/therealpaulgg/ssh-sync-server/test/pgx"
)

func TestUserRepo_DeleteUserKeyTx(t *testing.T) {
	injector := do.New()
	db := pgx.NewMockDatabase()

	// Mock transaction
	mockTx := &pgx.MockTx{}
	
	// Mock query service
	do.Provide(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshKey], error) {
		return &query.QueryServiceTxImpl[models.SshKey]{
			DataAccessor: db,
		}, nil
	})

	// Mock for change repository
	do.Provide(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshKeyChange], error) {
		return &query.QueryServiceTxImpl[models.SshKeyChange]{
			DataAccessor: db,
		}, nil
	})

	// Create repo under test
	repo := UserRepo{
		Injector: injector,
	}

	// Test data
	userID := uuid.New()
	keyID := uuid.New()
	user := &models.User{
		ID: userID,
	}
	
	// Set up mock query results
	db.MockTxQueryRow = func(tx pgx.Tx, sql string, args ...interface{}) []interface{} {
		if sql == "SELECT * FROM ssh_keys WHERE user_id = $1 AND id = $2" {
			assert.Equal(t, userID, args[0])
			assert.Equal(t, keyID, args[1])
			
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
			assert.Equal(t, models.Deleted, args[3])
			assert.Equal(t, "id_rsa.pub", args[4])
			assert.Equal(t, []byte("ssh-rsa AAAAB3NzaC..."), args[5])
			assert.Nil(t, args[6]) // No new data for deletion
			
			return []interface{}{
				args[0], // ID
				keyID,
				userID,
				models.Deleted,
				"id_rsa.pub",
				[]byte("ssh-rsa AAAAB3NzaC..."),
				nil,
				args[7], // ChangeTime
			}
		}
		return nil
	}
	
	// Mock the Exec call for deletion
	mockTx.ExecFunc = func(ctx context.Context, sql string, arguments ...interface{}) (pgx.CommandTag, error) {
		assert.Equal(t, "DELETE FROM ssh_keys WHERE user_id = $1 AND id = $2", sql)
		assert.Equal(t, userID, arguments[0])
		assert.Equal(t, keyID, arguments[1])
		
		return pgx.CommandTag("DELETE 1"), nil
	}

	// Run the test
	err := repo.DeleteUserKeyTx(user, keyID, mockTx)
	assert.NoError(t, err)
}
