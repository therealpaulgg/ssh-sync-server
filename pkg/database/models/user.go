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

var queryAccessor query.QueryThing[User] = &query.QueryImplementer[User]{}

func (u *User) GetUser(q query.QueryThing[User]) error {
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

func (u *User) GetUserByUsername(q query.QueryThing[User]) error {
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

func (u *User) CreateUser(q query.QueryThing[User]) error {
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
