package repository

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
	"github.com/therealpaulgg/ssh-sync-server/test/pgx"
)

func TestUpsertSshKeyTx(t *testing.T) {
	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tx := pgx.NewMockTx(ctrl)
	key := &models.SshKey{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		Filename: "id_rsa",
		Data:     []byte("ssh-rsa"),
	}

	mockQuery := query.NewMockQueryServiceTx[models.SshKey](ctrl)
	mockQuery.EXPECT().
		QueryOne(tx, gomock.Any(), key.UserID, key.Filename, key.Data).
		Return(key, nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryServiceTx[models.SshKey], error) {
		return mockQuery, nil
	})

	repo := &SshKeyRepo{Injector: injector}
	result, err := repo.UpsertSshKeyTx(key, tx)
	assert.NoError(t, err)
	assert.Equal(t, key, result)
}
