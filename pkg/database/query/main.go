package query

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database"
)

func Query[T any](query string, args ...interface{}) ([]T, error) {
	var results []T
	err := pgxscan.Select(context.Background(), database.DataAccessorInstance.GetConnection(), &results, query, args...)
	return results, err
}

func QueryOne[T any](query string, args ...interface{}) (*T, error) {
	rows, err := Query[T](query, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

func Insert[T any](query string, args ...interface{}) error {
	_, err := database.DataAccessorInstance.GetConnection().Exec(context.Background(), query, args...)
	return err
}
