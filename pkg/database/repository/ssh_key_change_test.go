package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
	testpgx "github.com/therealpaulgg/ssh-sync-server/test/pgx"
)

func setupSshKeyChangeRepoTest(t *testing.T) (*do.Injector, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	injector := do.New()
	return injector, ctrl
}

func TestSshKeyChangeRepo_CreateKeyChange(t *testing.T) {
	// Arrange
	injector, ctrl := setupSshKeyChangeRepoTest(t)
	defer ctrl.Finish()

	userID := uuid.New()
	keyID := uuid.New()
	changeID := uuid.New()
	now := time.Now().UTC()
	
	testChange := &models.SshKeyChange{
		ID:           changeID,
		SshKeyID:     keyID,
		UserID:       userID,
		ChangeType:   models.Created,
		Filename:     "test_key.pub",
		NewData:      []byte("ssh-rsa TEST"),
		ChangeTime:   now,
	}

	mockQuery := query.NewMockQueryService[models.SshKeyChange](ctrl)
	mockQuery.EXPECT().
		QueryOne(
			"INSERT INTO ssh_key_changes (id, ssh_key_id, user_id, change_type, filename, previous_data, new_data, change_time) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *",
			gomock.Any(), keyID, userID, models.Created, "test_key.pub", nil, []byte("ssh-rsa TEST"), gomock.Any(),
		).
		Return(testChange, nil)

	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.SshKeyChange], error) {
		return mockQuery, nil
	})

	repo := &SshKeyChangeRepo{Injector: injector}

	// Act
	change, err := repo.CreateKeyChange(testChange)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, testChange, change)
}

func TestSshKeyChangeRepo_CreateKeyChange_Error(t *testing.T) {
	// Arrange
	injector, ctrl := setupSshKeyChangeRepoTest(t)
	defer ctrl.Finish()

	userID := uuid.New()
	keyID := uuid.New()
	now := time.Now().UTC()
	
	testChange := &models.SshKeyChange{
		SshKeyID:     keyID,
		UserID:       userID,
		ChangeType:   models.Created,
		Filename:     "test_key.pub",
		NewData:      []byte("ssh-rsa TEST"),
		ChangeTime:   now,
	}

	mockQuery := query.NewMockQueryService[models.SshKeyChange](ctrl)
	mockQuery.EXPECT().
		QueryOne(
			"INSERT INTO ssh_key_changes (id, ssh_key_id, user_id, change_type, filename, previous_data, new_data, change_time) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *",
			gomock.Any(), keyID, userID, models.Created, "test_key.pub", nil, []byte("ssh-rsa TEST"), gomock.Any(),
		).
		Return(nil, errors.New("database error"))

	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.SshKeyChange], error) {
		return mockQuery, nil
	})

	repo := &SshKeyChangeRepo{Injector: injector}

	// Act
	change, err := repo.CreateKeyChange(testChange)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, change)
}

func TestSshKeyChangeRepo_CreateKeyChangeTx(t *testing.T) {
	// Arrange
	injector, ctrl := setupSshKeyChangeRepoTest(t)
	defer ctrl.Finish()

	userID := uuid.New()
	keyID := uuid.New()
	changeID := uuid.New()
	now := time.Now().UTC()
	
	testChange := &models.SshKeyChange{
		ID:           changeID,
		SshKeyID:     keyID,
		UserID:       userID,
		ChangeType:   models.Created,
		Filename:     "test_key.pub",
		NewData:      []byte("ssh-rsa TEST"),
		ChangeTime:   now,
	}

	tx := testpgx.NewMockTx(ctrl)
	mockQueryTx := query.NewMockQueryServiceTx[models.SshKeyChange](ctrl)
	mockQueryTx.EXPECT().
		QueryOne(
			tx,
			"INSERT INTO ssh_key_changes (id, ssh_key_id, user_id, change_type, filename, previous_data, new_data, change_time) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *",
			gomock.Any(), // id
			keyID,        // ssh_key_id
			userID,       // user_id
			models.Created, // change_type
			"test_key.pub", // filename
			nil,          // previous_data
			[]byte("ssh-rsa TEST"), // new_data
			gomock.Any(), // change_time
		).
		Return(testChange, nil)

	do.Provide(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshKeyChange], error) {
		return mockQueryTx, nil
	})

	repo := &SshKeyChangeRepo{Injector: injector}

	// Act
	change, err := repo.CreateKeyChangeTx(testChange, tx)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, testChange, change)
}

func TestSshKeyChangeRepo_GetKeyChanges(t *testing.T) {
	// Arrange
	injector, ctrl := setupSshKeyChangeRepoTest(t)
	defer ctrl.Finish()

	keyID := uuid.New()
	userID := uuid.New()
	
	changes := []models.SshKeyChange{
		{
			ID:         uuid.New(),
			SshKeyID:   keyID,
			UserID:     userID,
			ChangeType: models.Created,
			Filename:   "test_key.pub",
			NewData:    []byte("ssh-rsa TEST1"),
			ChangeTime: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:           uuid.New(),
			SshKeyID:     keyID,
			UserID:       userID,
			ChangeType:   models.Updated,
			Filename:     "test_key.pub",
			PreviousData: []byte("ssh-rsa TEST1"),
			NewData:      []byte("ssh-rsa TEST2"),
			ChangeTime:   time.Now().Add(-1 * time.Hour),
		},
	}

	mockQuery := query.NewMockQueryService[models.SshKeyChange](ctrl)
	mockQuery.EXPECT().
		Query(
			"SELECT * FROM ssh_key_changes WHERE ssh_key_id = $1 ORDER BY change_time DESC",
			keyID,
		).
		Return(changes, nil)

	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.SshKeyChange], error) {
		return mockQuery, nil
	})

	repo := &SshKeyChangeRepo{Injector: injector}

	// Act
	result, err := repo.GetKeyChanges(keyID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, changes, result)
}

func TestSshKeyChangeRepo_GetLatestKeyChangesForUser(t *testing.T) {
	// Arrange
	injector, ctrl := setupSshKeyChangeRepoTest(t)
	defer ctrl.Finish()

	userID := uuid.New()
	since := time.Now().Add(-24 * time.Hour)
	
	changes := []models.SshKeyChange{
		{
			ID:         uuid.New(),
			SshKeyID:   uuid.New(),
			UserID:     userID,
			ChangeType: models.Created,
			Filename:   "key1.pub",
			NewData:    []byte("ssh-rsa KEY1"),
			ChangeTime: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:           uuid.New(),
			SshKeyID:     uuid.New(),
			UserID:       userID,
			ChangeType:   models.Updated,
			Filename:     "key2.pub",
			PreviousData: []byte("ssh-rsa OLD"),
			NewData:      []byte("ssh-rsa KEY2"),
			ChangeTime:   time.Now().Add(-1 * time.Hour),
		},
	}

	mockQuery := query.NewMockQueryService[models.SshKeyChange](ctrl)
	mockQuery.EXPECT().
		Query(
			`SELECT DISTINCT ON (ssh_key_id) * 
		FROM ssh_key_changes 
		WHERE user_id = $1 AND change_time > $2
		ORDER BY ssh_key_id, change_time DESC`,
			userID, since,
		).
		Return(changes, nil)

	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.SshKeyChange], error) {
		return mockQuery, nil
	})

	repo := &SshKeyChangeRepo{Injector: injector}

	// Act
	result, err := repo.GetLatestKeyChangesForUser(userID, since)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, changes, result)
}

func TestSshKeyChangeMock(t *testing.T) {
	// Test the mock implementation
	mock := SshKeyChangeMock{}
	
	// Test CreateKeyChange
	userID := uuid.New()
	keyID := uuid.New()
	change := &models.SshKeyChange{
		SshKeyID:   keyID,
		UserID:     userID,
		ChangeType: models.Created,
		Filename:   "test.pub",
		NewData:    []byte("test data"),
	}
	
	result, err := mock.CreateKeyChange(change)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, result.ID)  // ID should be set
	assert.False(t, result.ChangeTime.IsZero()) // Time should be set
	
	// Test CreateKeyChangeTx - we just pass nil for the tx since it's not used in the mock
	result, err = mock.CreateKeyChangeTx(change, nil)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(mock.Changes)) // Should have added another change
	
	// Test GetKeyChanges
	changes, err := mock.GetKeyChanges(keyID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(changes)) // Should return both changes from above
	
	// Test GetLatestKeyChangesForUser
	since := time.Now().Add(-1 * time.Hour)
	latestChanges, err := mock.GetLatestKeyChangesForUser(userID, since)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(latestChanges)) // Should return one change per key
}