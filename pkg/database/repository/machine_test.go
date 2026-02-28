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

func TestCreateMachineAlreadyExists(t *testing.T) {
	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	machine := &models.Machine{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Name:      "duplicate",
		PublicKey: []byte("key"),
	}
	mockQuery := query.NewMockQueryService[models.Machine](ctrl)
	mockQuery.EXPECT().QueryOne("select * from machines where name = $1 and user_id = $2", machine.Name, machine.UserID).Return(machine, nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.Machine], error) {
		return mockQuery, nil
	})

	repo := &MachineRepo{Injector: injector}
	_, err := repo.CreateMachine(machine)
	assert.True(t, errors.Is(err, ErrMachineAlreadyExists))
}

func TestGetMachineNoRows(t *testing.T) {
	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	id := uuid.New()
	mockQuery := query.NewMockQueryService[models.Machine](ctrl)
	mockQuery.EXPECT().QueryOne("select * from machines where id = $1", id).Return(nil, nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.Machine], error) {
		return mockQuery, nil
	})

	repo := &MachineRepo{Injector: injector}
	machine, err := repo.GetMachine(id)
	assert.Nil(t, machine)
	assert.True(t, errors.Is(err, sql.ErrNoRows))
}

func TestGetMachineSuccess(t *testing.T) {
	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	id := uuid.New()
	expected := &models.Machine{ID: id, Name: "ok", UserID: uuid.New()}
	mockQuery := query.NewMockQueryService[models.Machine](ctrl)
	mockQuery.EXPECT().QueryOne("select * from machines where id = $1", id).Return(expected, nil)
	do.Provide(injector, func(i *do.Injector) (query.QueryService[models.Machine], error) {
		return mockQuery, nil
	})

	repo := &MachineRepo{Injector: injector}
	machine, err := repo.GetMachine(id)
	assert.NoError(t, err)
	assert.Equal(t, expected, machine)
}
