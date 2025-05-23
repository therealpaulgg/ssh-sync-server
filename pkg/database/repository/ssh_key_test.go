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

func setupSshKeyRepoTest(t *testing.T) (*do.Injector, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	injector := do.New()
	return injector, ctrl
}

func TestSshKeyRepo_CreateSshKey(t *testing.T) {
	// Arrange
	injector, ctrl := setupSshKeyRepoTest(t)
	defer ctrl.Finish()

	userId := uuid.New()
	keyId := uuid.New()
	testKey := &models.SshKey{
		ID:       keyId,
		UserID:   userId,
		Filename: "test_key.pub",
		Data:     []byte("ssh-rsa TEST"),
	}

	mockQuery := query.NewMockQueryService[models.SshKey](ctrl)
	mockQuery.EXPECT().
		QueryOne("INSERT INTO ssh_keys (user_id, filename, data) VALUES ($1, $2, $3) RETURNING *", 
			userId, "test_key.pub", []byte("ssh-rsa TEST")).
		Return(testKey, nil)

	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.SshKey], error) {
		return mockQuery, nil
	})

	repo := &SshKeyRepo{Injector: injector}

	// Act
	key, err := repo.CreateSshKey(testKey)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, testKey, key)
}

func TestSshKeyRepo_CreateSshKey_Error(t *testing.T) {
	// Arrange
	injector, ctrl := setupSshKeyRepoTest(t)
	defer ctrl.Finish()

	userId := uuid.New()
	testKey := &models.SshKey{
		UserID:   userId,
		Filename: "test_key.pub",
		Data:     []byte("ssh-rsa TEST"),
	}

	mockQuery := query.NewMockQueryService[models.SshKey](ctrl)
	mockQuery.EXPECT().
		QueryOne("INSERT INTO ssh_keys (user_id, filename, data) VALUES ($1, $2, $3) RETURNING *", 
			userId, "test_key.pub", []byte("ssh-rsa TEST")).
		Return(nil, errors.New("database error"))

	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.SshKey], error) {
		return mockQuery, nil
	})

	repo := &SshKeyRepo{Injector: injector}

	// Act
	key, err := repo.CreateSshKey(testKey)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, key)
}

func TestSshKeyRepo_UpsertSshKey(t *testing.T) {
	// Arrange
	injector, ctrl := setupSshKeyRepoTest(t)
	defer ctrl.Finish()

	userId := uuid.New()
	keyId := uuid.New()
	testKey := &models.SshKey{
		ID:       keyId,
		UserID:   userId,
		Filename: "test_key.pub",
		Data:     []byte("ssh-rsa TEST"),
	}

	mockQuery := query.NewMockQueryService[models.SshKey](ctrl)
	mockQuery.EXPECT().
		QueryOne("INSERT INTO ssh_keys (user_id, filename, data) VALUES ($1, $2, $3) ON CONFLICT (user_id, filename) DO UPDATE SET data = $3 RETURNING *", 
			userId, "test_key.pub", []byte("ssh-rsa TEST")).
		Return(testKey, nil)

	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.SshKey], error) {
		return mockQuery, nil
	})

	repo := &SshKeyRepo{Injector: injector}

	// Act
	key, err := repo.UpsertSshKey(testKey)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, testKey, key)
}

func TestSshKeyRepo_UpsertSshKeyTx(t *testing.T) {
	// Arrange
	injector, ctrl := setupSshKeyRepoTest(t)
	defer ctrl.Finish()

	userId := uuid.New()
	keyId := uuid.New()
	testKey := &models.SshKey{
		ID:       keyId,
		UserID:   userId,
		Filename: "test_key.pub",
		Data:     []byte("ssh-rsa TEST"),
	}

	tx := testpgx.NewMockTx(ctrl)
	mockQueryTx := query.NewMockQueryServiceTx[models.SshKey](ctrl)
	mockQueryTx.EXPECT().
		QueryOne(tx, "INSERT INTO ssh_keys (user_id, filename, data) VALUES ($1, $2, $3) ON CONFLICT (user_id, filename) DO UPDATE SET data = $3 RETURNING *", 
			userId, "test_key.pub", []byte("ssh-rsa TEST")).
		Return(testKey, nil)

	do.Provide(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshKey], error) {
		return mockQueryTx, nil
	})

	repo := &SshKeyRepo{Injector: injector}

	// Act
	key, err := repo.UpsertSshKeyTx(testKey, tx)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, testKey, key)
}

func TestSshKeyRepo_GetSshKeyByFilename(t *testing.T) {
	// Arrange
	injector, ctrl := setupSshKeyRepoTest(t)
	defer ctrl.Finish()

	userId := uuid.New()
	keyId := uuid.New()
	filename := "test_key.pub"
	testKey := &models.SshKey{
		ID:       keyId,
		UserID:   userId,
		Filename: filename,
		Data:     []byte("ssh-rsa TEST"),
	}

	mockQuery := query.NewMockQueryService[models.SshKey](ctrl)
	mockQuery.EXPECT().
		QueryOne("SELECT * FROM ssh_keys WHERE user_id = $1 AND filename = $2", userId, filename).
		Return(testKey, nil)

	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.SshKey], error) {
		return mockQuery, nil
	})

	repo := &SshKeyRepo{Injector: injector}

	// Act
	key, err := repo.GetSshKeyByFilename(userId, filename)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, testKey, key)
}

func TestSshKeyRepo_CreateSshKeyWithChange(t *testing.T) {
	// Arrange
	injector, ctrl := setupSshKeyRepoTest(t)
	defer ctrl.Finish()

	userId := uuid.New()
	keyId := uuid.New()
	changeId := uuid.New()
	testKey := &models.SshKey{
		ID:       keyId,
		UserID:   userId,
		Filename: "test_key.pub",
		Data:     []byte("ssh-rsa TEST"),
	}

	testChange := &models.SshKeyChange{
		ID:         changeId,
		SshKeyID:   keyId,
		UserID:     userId,
		ChangeType: models.Created,
		Filename:   "test_key.pub",
		NewData:    []byte("ssh-rsa TEST"),
		ChangeTime: time.Now(),
	}

	tx := testpgx.NewMockTx(ctrl)
	
	// Mock transaction service
	mockTxService := query.NewMockTransactionService(ctrl)
	mockTxService.EXPECT().StartTx(gomock.Any()).Return(tx, nil)
	mockTxService.EXPECT().Commit(tx).Return(nil)
	do.Provide(injector, func(i *do.Injector) (query.TransactionService, error) {
		return mockTxService, nil
	})
	
	// Mock query services for UpsertSshKeyTx
	mockQueryTx := query.NewMockQueryServiceTx[models.SshKey](ctrl)
	mockQueryTx.EXPECT().
		QueryOne(gomock.Eq(tx), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(testKey, nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshKey], error) {
		return mockQueryTx, nil
	})
	
	// Mock query services for CreateKeyChangeTx
	mockChangeTx := query.NewMockQueryServiceTx[models.SshKeyChange](ctrl)
	mockChangeTx.EXPECT().
		QueryOne(
			gomock.Eq(tx), 
			gomock.Any(),
			gomock.Any(), // id
			gomock.Any(), // ssh_key_id
			gomock.Any(), // user_id
			gomock.Any(), // change_type
			gomock.Any(), // filename
			gomock.Any(), // previous_data
			gomock.Any(), // new_data
			gomock.Any(), // change_time
		).
		Return(testChange, nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshKeyChange], error) {
		return mockChangeTx, nil
	})

	repo := &SshKeyRepo{Injector: injector}

	// Act
	key, err := repo.CreateSshKeyWithChange(testKey)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, testKey, key)
}

func TestSshKeyRepo_UpsertSshKeyWithChange(t *testing.T) {
	// Arrange
	injector, ctrl := setupSshKeyRepoTest(t)
	defer ctrl.Finish()

	userId := uuid.New()
	keyId := uuid.New()
	changeId := uuid.New()
	testKey := &models.SshKey{
		ID:       keyId,
		UserID:   userId,
		Filename: "test_key.pub",
		Data:     []byte("ssh-rsa TEST"),
	}
	
	existingKey := &models.SshKey{
		ID:       keyId,
		UserID:   userId,
		Filename: "test_key.pub",
		Data:     []byte("ssh-rsa OLD"),
	}

	testChange := &models.SshKeyChange{
		ID:           changeId,
		SshKeyID:     keyId,
		UserID:       userId,
		ChangeType:   models.Updated,
		Filename:     "test_key.pub",
		PreviousData: []byte("ssh-rsa OLD"),
		NewData:      []byte("ssh-rsa TEST"),
		ChangeTime:   time.Now(),
	}

	tx := testpgx.NewMockTx(ctrl)
	
	// Mock transaction service
	mockTxService := query.NewMockTransactionService(ctrl)
	mockTxService.EXPECT().StartTx(gomock.Any()).Return(tx, nil)
	mockTxService.EXPECT().Commit(tx).Return(nil)
	do.Provide(injector, func(i *do.Injector) (query.TransactionService, error) {
		return mockTxService, nil
	})
	
	// Mock query service for GetSshKeyByFilename
	mockQuery := query.NewMockQueryService[models.SshKey](ctrl)
	mockQuery.EXPECT().
		QueryOne("SELECT * FROM ssh_keys WHERE user_id = $1 AND filename = $2", userId, "test_key.pub").
		Return(existingKey, nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.SshKey], error) {
		return mockQuery, nil
	})
	
	// Mock query services for UpsertSshKeyTx
	mockQueryTx := query.NewMockQueryServiceTx[models.SshKey](ctrl)
	mockQueryTx.EXPECT().
		QueryOne(gomock.Eq(tx), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(testKey, nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshKey], error) {
		return mockQueryTx, nil
	})
	
	// Mock query services for CreateKeyChangeTx
	mockChangeTx := query.NewMockQueryServiceTx[models.SshKeyChange](ctrl)
	mockChangeTx.EXPECT().
		QueryOne(
			gomock.Eq(tx), 
			gomock.Any(),
			gomock.Any(), // id
			gomock.Any(), // ssh_key_id
			gomock.Any(), // user_id
			gomock.Any(), // change_type
			gomock.Any(), // filename
			gomock.Any(), // previous_data
			gomock.Any(), // new_data
			gomock.Any(), // change_time
		).
		Return(testChange, nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshKeyChange], error) {
		return mockChangeTx, nil
	})

	repo := &SshKeyRepo{Injector: injector}

	// Act
	key, err := repo.UpsertSshKeyWithChange(testKey)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, testKey, key)
}

func TestSshKeyRepo_UpsertSshKeyWithChangeTx(t *testing.T) {
	// Arrange
	injector, ctrl := setupSshKeyRepoTest(t)
	defer ctrl.Finish()

	userId := uuid.New()
	keyId := uuid.New()
	changeId := uuid.New()
	testKey := &models.SshKey{
		ID:       keyId,
		UserID:   userId,
		Filename: "test_key.pub",
		Data:     []byte("ssh-rsa TEST"),
	}
	
	existingKey := &models.SshKey{
		ID:       keyId,
		UserID:   userId,
		Filename: "test_key.pub",
		Data:     []byte("ssh-rsa OLD"),
	}

	testChange := &models.SshKeyChange{
		ID:           changeId,
		SshKeyID:     keyId,
		UserID:       userId,
		ChangeType:   models.Updated,
		Filename:     "test_key.pub",
		PreviousData: []byte("ssh-rsa OLD"),
		NewData:      []byte("ssh-rsa TEST"),
		ChangeTime:   time.Now(),
	}

	tx := testpgx.NewMockTx(ctrl)
	
	// Mock query service for GetSshKeyByFilename
	mockQuery := query.NewMockQueryService[models.SshKey](ctrl)
	mockQuery.EXPECT().
		QueryOne("SELECT * FROM ssh_keys WHERE user_id = $1 AND filename = $2", userId, "test_key.pub").
		Return(existingKey, nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.SshKey], error) {
		return mockQuery, nil
	})
	
	// Mock query services for UpsertSshKeyTx
	mockQueryTx := query.NewMockQueryServiceTx[models.SshKey](ctrl)
	mockQueryTx.EXPECT().
		QueryOne(gomock.Eq(tx), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(testKey, nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshKey], error) {
		return mockQueryTx, nil
	})
	
	// Mock query services for CreateKeyChangeTx
	mockChangeTx := query.NewMockQueryServiceTx[models.SshKeyChange](ctrl)
	mockChangeTx.EXPECT().
		QueryOne(
			gomock.Eq(tx), 
			gomock.Any(),
			gomock.Any(), // id
			gomock.Any(), // ssh_key_id
			gomock.Any(), // user_id
			gomock.Any(), // change_type
			gomock.Any(), // filename
			gomock.Any(), // previous_data
			gomock.Any(), // new_data
			gomock.Any(), // change_time
		).
		Return(testChange, nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshKeyChange], error) {
		return mockChangeTx, nil
	})

	repo := &SshKeyRepo{Injector: injector}

	// Act
	key, err := repo.UpsertSshKeyWithChangeTx(testKey, tx)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, testKey, key)
}
