package repository

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
)

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
