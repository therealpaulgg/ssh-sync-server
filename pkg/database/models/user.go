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

type User struct {
	ID       uuid.UUID   `json:"id" db:"id"`
	Username string      `json:"username" db:"username"`
	Keys     []SshKey    `json:"keys"`
	Config   []SshConfig `json:"config"`
	Machines []Machine   `json:"machines"`
}

var ErrUserAlreadyExists = errors.New("user already exists")

func (u *User) GetUser(i *do.Injector) error {
	q := do.MustInvoke[query.QueryService[User]](i)
	user, err := q.QueryOne("select * from users where id = $1", u.ID)
	if err != nil {
		return err
	}
	if user == nil {
		return sql.ErrNoRows
	}
	u.ID = user.ID
	u.Username = user.Username
	return nil
}

func (u *User) GetUserByUsername(i *do.Injector) error {
	q := do.MustInvoke[query.QueryService[User]](i)
	user, err := q.QueryOne("select * from users where username = $1", u.Username)
	if err != nil {
		return err
	}
	if user == nil {
		return sql.ErrNoRows
	}
	u.ID = user.ID
	u.Username = user.Username
	return nil
}

func (u *User) CreateUser(i *do.Injector) error {
	q := do.MustInvoke[query.QueryService[User]](i)
	existingUser, err := q.QueryOne("select * from users where username = $1", u.Username)
	if err != nil {
		return err
	}

	if existingUser != nil {
		return ErrUserAlreadyExists
	}
	user, err := q.QueryOne("insert into users (username) values ($1) returning *", u.Username)
	if err != nil {
		return err
	}
	if user == nil {
		return sql.ErrNoRows
	}
	u.ID = user.ID
	u.Username = user.Username
	return nil
}

func (u *User) CreateUserTx(i *do.Injector, tx pgx.Tx) error {
	q := do.MustInvoke[query.QueryServiceTx[User]](i)
	existingUser, err := q.QueryOne(tx, "select * from users where username = $1", u.Username)
	if err != nil {
		return err
	}

	if existingUser != nil {
		return ErrUserAlreadyExists
	}
	user, err := q.QueryOne(tx, "insert into users (username) values ($1) returning *", u.Username)
	if err != nil {
		return err
	}
	if user == nil {
		return sql.ErrNoRows
	}
	u.ID = user.ID
	u.Username = user.Username
	return nil
}

func (u *User) DeleteUser(i *do.Injector) error {
	q := do.MustInvoke[database.DataAccessor](i)
	tx, err := q.GetConnection().BeginTx(context.TODO(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	if _, err := tx.Exec(context.TODO(), "delete from ssh_keys where user_id = $1", u.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(context.TODO(), "delete from ssh_configs where user_id = $1", u.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(context.TODO(), "delete from master_keys where user_id = $1", u.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(context.TODO(), "delete from machines where user_id = $1", u.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(context.TODO(), "delete from users where id = $1", u.ID); err != nil {
		return err
	}
	if err := tx.Commit(context.TODO()); err != nil && !errors.Is(err, pgx.ErrTxCommitRollback) {
		return tx.Rollback(context.TODO())
	}
	return nil
}

func (u *User) GetUserConfig(i *do.Injector) error {
	q := do.MustInvoke[query.QueryService[SshConfig]](i)
	config, err := q.Query("select * from ssh_configs where user_id = $1", u.ID)
	if err != nil {
		return err
	}
	u.Config = config
	return nil
}

func (u *User) GetUserKeys(i *do.Injector) error {
	q := do.MustInvoke[query.QueryService[SshKey]](i)
	keys, err := q.Query("select * from ssh_keys where user_id = $1", u.ID)
	if err != nil {
		return err
	}
	u.Keys = keys
	return nil
}

func (u *User) GetUserMachines(i *do.Injector) error {
	q := do.MustInvoke[query.QueryService[Machine]](i)
	machines, err := q.Query("select * from machines where user_id = $1", u.ID)
	if err != nil {
		return err
	}
	u.Machines = machines
	return nil
}

func (u *User) AddAndUpdateKeys(i *do.Injector) error {
	for _, key := range u.Keys {
		key.UserID = u.ID
		err := key.UpsertSshKey(i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *User) AddAndUpdateConfig(i *do.Injector) error {
	for _, config := range u.Config {
		err := config.UpsertSshConfig(i)
		if err != nil {
			return err
		}
	}
	return nil
}
