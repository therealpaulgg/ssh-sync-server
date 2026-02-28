package repository

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
)

type stubConfigRepo struct {
	err error
}

func (s stubConfigRepo) GetSshConfig(uuid.UUID) (*models.SshConfig, error) {
	return nil, s.err
}

func (s stubConfigRepo) UpsertSshConfig(*models.SshConfig) (*models.SshConfig, error) {
	return nil, s.err
}

func (s stubConfigRepo) UpsertSshConfigTx(*models.SshConfig, pgx.Tx) (*models.SshConfig, error) {
	return nil, s.err
}

func TestCreateUserAlreadyExists(t *testing.T) {
	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	username := "existing"
	mockQuery := query.NewMockQueryService[models.User](ctrl)
	mockQuery.EXPECT().QueryOne("select * from users where username = $1", username).Return(&models.User{
		ID:       uuid.New(),
		Username: username,
	}, nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.User], error) {
		return mockQuery, nil
	})

	repo := &UserRepo{Injector: injector}
	_, err := repo.CreateUser(&models.User{Username: username})
	assert.ErrorIs(t, err, ErrUserAlreadyExists)
}

func TestGetUserNoRows(t *testing.T) {
	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := uuid.New()
	mockQuery := query.NewMockQueryService[models.User](ctrl)
	mockQuery.EXPECT().QueryOne("select * from users where id = $1", userID).Return(nil, nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.User], error) {
		return mockQuery, nil
	})

	repo := &UserRepo{Injector: injector}
	user, err := repo.GetUser(userID)
	assert.Nil(t, user)
	assert.True(t, errors.Is(err, sql.ErrNoRows))
}

func TestAddAndUpdateKeysError(t *testing.T) {
	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	user := &models.User{
		ID: uuid.New(),
		Keys: []models.SshKey{
			{ID: uuid.New(), UserID: uuid.New(), Filename: "id_rsa"},
		},
	}
	mockKeyRepo := NewMockSshKeyRepository(ctrl)
	mockKeyRepo.EXPECT().UpsertSshKey(gomock.Any()).Return(nil, errors.New("failure"))
	do.Provide(injector, func(i *do.Injector) (SshKeyRepository, error) {
		return mockKeyRepo, nil
	})

	repo := &UserRepo{Injector: injector}
	err := repo.AddAndUpdateKeys(user)
	assert.Error(t, err)
}

func TestAddAndUpdateConfigError(t *testing.T) {
	injector := do.New()

	user := &models.User{
		ID: uuid.New(),
		Config: []models.SshConfig{
			{UserID: uuid.New(), Host: "test"},
		},
	}
	do.Provide(injector, func(i *do.Injector) (SshConfigRepository, error) {
		return stubConfigRepo{err: errors.New("failure")}, nil
	})

	repo := &UserRepo{Injector: injector}
	err := repo.AddAndUpdateConfig(user)
	assert.Error(t, err)
}
