package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
	"github.com/therealpaulgg/ssh-sync-server/test/pgx"
)

func TestSshKeyChangeRepo_CreateKeyChange(t *testing.T) {
	injector := do.New()
	db := pgx.NewMockDatabase()
	
	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.SshKeyChange], error) {
		return &query.QueryServiceImpl[models.SshKeyChange]{
			DataAccessor: db,
		}, nil
	})
	
	repo := SshKeyChangeRepo{
		Injector: injector,
	}
	
	keyID := uuid.New()
	userID := uuid.New()
	
	change := &models.SshKeyChange{
		SshKeyID:     keyID,
		UserID:       userID,
		ChangeType:   models.Created,
		Filename:     "id_rsa",
		NewData:      []byte("ssh-key-data"),
		ChangeTime:   time.Now(),
	}
	
	// Set up mock response
	db.MockQueryRow = func(sql string, args ...interface{}) []interface{} {
		assert.Contains(t, sql, "INSERT INTO ssh_key_changes")
		return []interface{}{
			change.ID,
			change.SshKeyID,
			change.UserID,
			change.ChangeType,
			change.Filename,
			change.PreviousData,
			change.NewData,
			change.ChangeTime,
		}
	}
	
	result, err := repo.CreateKeyChange(change)
	
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, change.SshKeyID, result.SshKeyID)
	assert.Equal(t, change.UserID, result.UserID)
	assert.Equal(t, change.ChangeType, result.ChangeType)
	assert.Equal(t, change.Filename, result.Filename)
	assert.Equal(t, change.NewData, result.NewData)
}

func TestSshKeyChangeRepo_GetKeyChanges(t *testing.T) {
	injector := do.New()
	db := pgx.NewMockDatabase()
	
	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.SshKeyChange], error) {
		return &query.QueryServiceImpl[models.SshKeyChange]{
			DataAccessor: db,
		}, nil
	})
	
	repo := SshKeyChangeRepo{
		Injector: injector,
	}
	
	keyID := uuid.New()
	userID := uuid.New()
	
	now := time.Now()
	earlier := now.Add(-1 * time.Hour)
	
	expected := []models.SshKeyChange{
		{
			ID:          uuid.New(),
			SshKeyID:    keyID,
			UserID:      userID,
			ChangeType:  models.Updated,
			Filename:    "id_rsa",
			NewData:     []byte("updated-data"),
			ChangeTime:  now,
		},
		{
			ID:          uuid.New(),
			SshKeyID:    keyID,
			UserID:      userID,
			ChangeType:  models.Created,
			Filename:    "id_rsa",
			NewData:     []byte("original-data"),
			ChangeTime:  earlier,
		},
	}
	
	// Set up mock response
	db.MockQuery = func(sql string, args ...interface{}) [][]interface{} {
		assert.Contains(t, sql, "SELECT * FROM ssh_key_changes WHERE ssh_key_id = $1")
		assert.Equal(t, keyID, args[0])
		
		result := [][]interface{}{}
		for _, change := range expected {
			result = append(result, []interface{}{
				change.ID,
				change.SshKeyID,
				change.UserID,
				change.ChangeType,
				change.Filename,
				change.PreviousData,
				change.NewData,
				change.ChangeTime,
			})
		}
		
		return result
	}
	
	changes, err := repo.GetKeyChanges(keyID)
	
	assert.NoError(t, err)
	assert.Len(t, changes, 2)
	assert.Equal(t, expected[0].ID, changes[0].ID)
	assert.Equal(t, expected[1].ID, changes[1].ID)
}