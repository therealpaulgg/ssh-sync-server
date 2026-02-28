package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
)

type MasterKeyRotationRepository interface {
	UpsertRotation(machineID uuid.UUID, encKey []byte) error
	GetRotationForMachine(machineID uuid.UUID) (*models.MasterKeyRotation, error)
	DeleteRotationForMachine(machineID uuid.UUID) error
}

type MasterKeyRotationRepo struct {
	Injector *do.Injector
}

func (repo *MasterKeyRotationRepo) UpsertRotation(machineID uuid.UUID, encKey []byte) error {
	q := do.MustInvoke[database.DataAccessor](repo.Injector)
	_, err := q.GetConnection().Exec(
		context.TODO(),
		`INSERT INTO master_key_rotations (machine_id, encrypted_master_key)
		 VALUES ($1, $2)
		 ON CONFLICT (machine_id) DO UPDATE SET encrypted_master_key = EXCLUDED.encrypted_master_key, created_at = now() AT TIME ZONE 'UTC'`,
		machineID, encKey,
	)
	return err
}

func (repo *MasterKeyRotationRepo) GetRotationForMachine(machineID uuid.UUID) (*models.MasterKeyRotation, error) {
	q := do.MustInvoke[query.QueryService[models.MasterKeyRotation]](repo.Injector)
	rotation, err := q.QueryOne("SELECT * FROM master_key_rotations WHERE machine_id = $1", machineID)
	if err != nil {
		return nil, err
	}
	if rotation == nil {
		return nil, sql.ErrNoRows
	}
	return rotation, nil
}

func (repo *MasterKeyRotationRepo) DeleteRotationForMachine(machineID uuid.UUID) error {
	q := do.MustInvoke[database.DataAccessor](repo.Injector)
	_, err := q.GetConnection().Exec(
		context.TODO(),
		"DELETE FROM master_key_rotations WHERE machine_id = $1",
		machineID,
	)
	return err
}

var _ MasterKeyRotationRepository = (*MasterKeyRotationRepo)(nil)
