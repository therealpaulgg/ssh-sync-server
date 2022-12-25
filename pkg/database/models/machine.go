package models

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
)

type Machine struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	PublicKey []byte    `json:"public_key" db:"public_key"`
}

var ErrMachineAlreadyExists = errors.New("machine w/ user already exists")

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
