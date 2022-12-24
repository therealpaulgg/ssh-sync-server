package query

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database"
)

type QueryThing[T any] interface {
	Query(query string, args ...interface{}) ([]T, error)
	QueryOne(query string, args ...interface{}) (*T, error)
	Insert(query string, args ...interface{}) error
}

type QueryImplementer[T any] struct {
	DataAccessor database.DataAccessor
}

func (q *QueryImplementer[T]) Query(query string, args ...interface{}) ([]T, error) {
	var results []T
	err := pgxscan.Select(context.Background(), q.DataAccessor.GetConnection(), &results, query, args...)
	return results, err
}

func (q *QueryImplementer[T]) QueryOne(query string, args ...interface{}) (*T, error) {
	rows, err := q.Query(query, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

func (q *QueryImplementer[T]) Insert(query string, args ...interface{}) error {
	_, err := q.DataAccessor.GetConnection().Exec(context.Background(), query, args...)
	return err
}
