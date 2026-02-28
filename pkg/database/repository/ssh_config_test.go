package repository

import (
	"database/sql"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
)

func TestGetSshConfigNoRows(t *testing.T) {
	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := uuid.New()
	mockQuery := query.NewMockQueryService[models.SshConfig](ctrl)
	mockQuery.EXPECT().QueryOne("select * from ssh_configs where user_id = $1", userID).Return(nil, nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.SshConfig], error) {
		return mockQuery, nil
	})

	repo := &SshConfigRepo{Injector: injector}
	config, err := repo.GetSshConfig(userID)
	assert.Nil(t, config)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}
