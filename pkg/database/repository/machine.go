package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
)

type MachineRepository interface {
	DeleteMachine(id uuid.UUID) error
	GetMachine(id uuid.UUID) (*models.Machine, error)
	GetMachineByNameAndUser(machineName string, userID uuid.UUID) (*models.Machine, error)
	CreateMachine(machine *models.Machine) (*models.Machine, error)
	CreateMachineTx(machine *models.Machine, tx pgx.Tx) (*models.Machine, error)
	GetUserMachines(id uuid.UUID) ([]models.Machine, error)
}

type MachineRepo struct {
	Injector *do.Injector
}

var ErrMachineAlreadyExists = errors.New("machine w/ user already exists")

func (repo *MachineRepo) DeleteMachine(id uuid.UUID) error {
	q := do.MustInvoke[database.DataAccessor](repo.Injector)
	tx, err := q.GetConnection().BeginTx(context.TODO(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil && !errors.Is(err, pgx.ErrTxCommitRollback) {
			tx.Rollback(context.TODO())
		}
	}()
	if _, err = tx.Exec(context.TODO(), "delete from machines where id = $1", id); err != nil {
		return err
	}
	return tx.Commit(context.TODO())
}

func (repo *MachineRepo) GetMachine(id uuid.UUID) (*models.Machine, error) {
	q := do.MustInvoke[query.QueryService[models.Machine]](repo.Injector)
	machine, err := q.QueryOne("select * from machines where id = $1", id)
	if err != nil {
		return nil, err
	}
	if machine == nil {
		return nil, sql.ErrNoRows
	}
	return machine, nil
}

func (repo *MachineRepo) GetMachineByNameAndUser(machineName string, userID uuid.UUID) (*models.Machine, error) {
	q := do.MustInvoke[query.QueryService[models.Machine]](repo.Injector)
	machine, err := q.QueryOne("select * from machines where name = $1 and user_id = $2", machineName, userID)
	if err != nil {
		return nil, err
	}
	if machine == nil {
		return nil, sql.ErrNoRows
	}
	return machine, nil
}

func (repo *MachineRepo) CreateMachine(machine *models.Machine) (*models.Machine, error) {
	q := do.MustInvoke[query.QueryService[models.Machine]](repo.Injector)
	existingMachine, err := q.QueryOne("select * from machines where name = $1 and user_id = $2", machine.Name, machine.UserID)
	if err != nil {
		return nil, err
	}
	if existingMachine != nil {
		return nil, ErrMachineAlreadyExists
	}
	newMachine, err := q.QueryOne("insert into machines (user_id, name, public_key) values ($1, $2, $3) returning *", machine.UserID, machine.Name, machine.PublicKey)
	if err != nil {
		return nil, err
	}
	if newMachine == nil {
		return nil, sql.ErrNoRows
	}
	return newMachine, nil
}

func (repo *MachineRepo) CreateMachineTx(machine *models.Machine, tx pgx.Tx) (*models.Machine, error) {
	q := do.MustInvoke[query.QueryServiceTx[models.Machine]](repo.Injector)
	existingMachine, err := q.QueryOne(tx, "select * from machines where name = $1 and user_id = $2", machine.Name, machine.UserID)
	if err != nil {
		return nil, err
	}
	if existingMachine != nil {
		return nil, ErrMachineAlreadyExists
	}
	newMachine, err := q.QueryOne(tx, "insert into machines (user_id, name, public_key) values ($1, $2, $3) returning *", machine.UserID, machine.Name, machine.PublicKey)
	if err != nil {
		return nil, err
	}
	if newMachine == nil {
		return nil, sql.ErrNoRows
	}
	return newMachine, nil
}

func (repo *MachineRepo) GetUserMachines(id uuid.UUID) ([]models.Machine, error) {
	q := do.MustInvoke[query.QueryService[models.Machine]](repo.Injector)
	machines, err := q.Query("select * from machines where user_id = $1", id)
	if err != nil {
		return nil, err
	}
	return machines, nil
}
