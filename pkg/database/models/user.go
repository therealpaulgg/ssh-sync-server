package models

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
)

type User struct {
	ID       uuid.UUID `json:"id" db:"id"`
	Username string    `json:"username" db:"username"`
	Keys     []SshKey  `json:"keys"`
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

func (u *User) DeleteUser(i *do.Injector) error {
	q := do.MustInvoke[query.QueryService[User]](i)
	_, err := q.QueryOne("delete from users where id = $1", u.ID)
	return err
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
