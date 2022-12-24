package database

import (
	"context"
	"net/url"
	"os"

	"github.com/jackc/pgx/v5"
)

type DataAccessor interface {
	Connect() error
	GetConnection() *pgx.Conn
}

type DataAccessorImpl struct {
	Connection *pgx.Conn
}

func (d *DataAccessorImpl) Connect() error {
	data := url.URL{
		Scheme: "postgresql",
		User:   url.UserPassword(os.Getenv("DATABASE_USERNAME"), os.Getenv("DATABASE_PASSWORD")),
		Host:   os.Getenv("DATABASE_HOST"),
	}
	data.Query().Add("database", os.Getenv("DATABASE_NAME"))
	conn, err := pgx.Connect(context.Background(), data.String())
	if err != nil {
		return err
	}
	d.Connection = conn
	return nil
}

func (d *DataAccessorImpl) GetConnection() *pgx.Conn {
	return d.Connection
}

var DataAccessorInstance DataAccessor = &DataAccessorImpl{
	Connection: nil,
}
