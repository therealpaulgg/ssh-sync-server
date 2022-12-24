package models

import (
	"database/sql"
	"errors"

	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
)

type User struct {
	ID       string `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
}

var ErrUserAlreadyExists = errors.New("user already exists")

func (u *User) GetUser() error {
	user, err := query.QueryOne[User]("select * from users where id = $1", u.ID)
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

func (u *User) GetUserByUsername() error {
	user, err := query.QueryOne[User]("select * from users where username = $1", u.Username)
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

func (u *User) CreateUser() error {
	existingUser, err := query.QueryOne[User]("select * from users where username = $1", u.Username)
	if err != nil {
		return err
	}

	if existingUser != nil {
		return ErrUserAlreadyExists
	}
	user, err := query.QueryOne[User]("insert into users (username) values ($1) returning *", u.Username)
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
