package models

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
)

type Machine struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	PublicKey []byte    `json:"public_key" db:"public_key"`
}

var ErrMachineAlreadyExists = errors.New("machine w/ user already exists")

func (m *Machine) DeleteMachine(i *do.Injector) error {
	q := do.MustInvoke[database.DataAccessor](i)
	tx, err := q.GetConnection().BeginTx(context.TODO(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil && !errors.Is(err, pgx.ErrTxCommitRollback) {
			tx.Rollback(context.TODO())
		}
	}()
	if _, err := tx.Exec(context.TODO(), "delete from ssh_configs where machine_id = $1", m.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(context.TODO(), "delete from machines where id = $1", m.ID); err != nil {
		return err
	}
	err = tx.Commit(context.TODO())
	return err
}

func (m *Machine) GetMachine(i *do.Injector) error {
	q := do.MustInvoke[query.QueryService[Machine]](i)
	machine, err := q.QueryOne("select * from machines where id = $1", m.ID)
	if err != nil {
		return err
	}
	if machine == nil {
		return sql.ErrNoRows
	}
	m.ID = machine.ID
	m.Name = machine.Name
	m.UserID = machine.UserID
	m.PublicKey = machine.PublicKey
	return nil
}

func (m *Machine) GetMachineByNameAndUser(i *do.Injector) error {
	q := do.MustInvoke[query.QueryService[Machine]](i)
	machine, err := q.QueryOne("select * from machines where name = $1 and user_id = $2", m.Name, m.UserID)
	if err != nil {
		return err
	}
	if machine == nil {
		return sql.ErrNoRows
	}
	m.ID = machine.ID
	m.Name = machine.Name
	m.UserID = machine.UserID
	m.PublicKey = machine.PublicKey
	return nil
}

func (m *Machine) CreateMachine(i *do.Injector) error {
	q := do.MustInvoke[query.QueryService[Machine]](i)
	existingMachine, err := q.QueryOne("select * from machines where name = $1 and user_id = $2", m.Name, m.UserID)
	if err != nil {
		return err
	}
	if existingMachine != nil {
		return ErrMachineAlreadyExists
	}
	machine, err := q.QueryOne("insert into machines (user_id, name, public_key) values ($1, $2, $3) returning *", m.UserID, m.Name, m.PublicKey)
	if err != nil {
		return err
	}
	if machine == nil {
		return sql.ErrNoRows
	}
	m.ID = machine.ID
	m.Name = machine.Name
	m.UserID = machine.UserID
	m.PublicKey = machine.PublicKey
	return nil
}

func (m *Machine) CreateMachineTx(i *do.Injector, tx pgx.Tx) error {
	q := do.MustInvoke[query.QueryServiceTx[Machine]](i)
	existingMachine, err := q.QueryOne(tx, "select * from machines where name = $1 and user_id = $2", m.Name, m.UserID)
	if err != nil {
		return err
	}
	if existingMachine != nil {
		return ErrMachineAlreadyExists
	}
	machine, err := q.QueryOne(tx, "insert into machines (user_id, name, public_key) values ($1, $2, $3) returning *", m.UserID, m.Name, m.PublicKey)
	if err != nil {
		return err
	}
	if machine == nil {
		return sql.ErrNoRows
	}
	m.ID = machine.ID
	m.Name = machine.Name
	m.UserID = machine.UserID
	m.PublicKey = machine.PublicKey
	return nil
}
