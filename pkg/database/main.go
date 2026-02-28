package database

//go:generate go run go.uber.org/mock/mockgen -destination=mock.go -package=database github.com/therealpaulgg/ssh-sync-server/pkg/database DataAccessor

import (
	"context"
	"net/url"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/samber/do"
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

func NewDataAccessorService(i *do.Injector) (DataAccessor, error) {
	accessor := &DataAccessorImpl{
		Connection: nil,
	}
	err := accessor.Connect()
	return accessor, err
}

func (d *DataAccessorImpl) GetConnection() *pgx.Conn {
	return d.Connection
}
