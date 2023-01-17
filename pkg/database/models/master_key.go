package models

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
)

type MasterKey struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	MachineID uuid.UUID `json:"machine_id" db:"machine_id"`
	Data      []byte    `json:"data" db:"data"`
}

var ErrKeyAlreadyExists = errors.New("key already exists")

func (m *MasterKey) GetMasterKey(i *do.Injector) error {
	q := do.MustInvoke[query.QueryService[MasterKey]](i)
	masterKey, err := q.QueryOne("select * from master_keys where machine_id = $1 and user_id = $2", m.MachineID, m.UserID)
	if err != nil {
		return err
	}
	if masterKey == nil {
		return sql.ErrNoRows
	}
	m.ID = masterKey.ID
	m.UserID = masterKey.UserID
	m.MachineID = masterKey.MachineID
	m.Data = masterKey.Data
	return nil
}

func (m *MasterKey) CreateMasterKey(i *do.Injector) error {
	q := do.MustInvoke[query.QueryService[MasterKey]](i)
	existingKey, err := q.QueryOne("select * from master_keys where machine_id = $1 and user_id = $2", m.MachineID, m.UserID)
	if err != nil {
		return err
	}

	if existingKey != nil {
		return ErrKeyAlreadyExists
	}
	masterKey, err := q.QueryOne("insert into master_keys (user_id, machine_id, data) values ($1, $2, $3) returning *", m.UserID, m.MachineID, m.Data)
	if err != nil {
		return err
	}
	if masterKey == nil {
		return sql.ErrNoRows
	}
	m.ID = masterKey.ID
	m.UserID = masterKey.UserID
	m.MachineID = masterKey.MachineID
	m.Data = masterKey.Data
	return nil
}

func (m *MasterKey) CreateMasterKeyTx(i *do.Injector, tx *pgx.Tx) error {
	q := do.MustInvoke[query.QueryServiceTx[MasterKey]](i)
	existingKey, err := q.QueryOne(tx, "select * from master_keys where machine_id = $1 and user_id = $2", m.MachineID, m.UserID)
	if err != nil {
		return err
	}

	if existingKey != nil {
		return ErrKeyAlreadyExists
	}
	masterKey, err := q.QueryOne(tx, "insert into master_keys (user_id, machine_id, data) values ($1, $2, $3) returning *", m.UserID, m.MachineID, m.Data)
	if err != nil {
		return err
	}
	if masterKey == nil {
		return sql.ErrNoRows
	}
	m.ID = masterKey.ID
	m.UserID = masterKey.UserID
	m.MachineID = masterKey.MachineID
	m.Data = masterKey.Data
	return nil
}
