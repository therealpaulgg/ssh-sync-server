package query

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database"
)

type QueryService[T any] interface {
	Query(query string, args ...any) ([]T, error)
	QueryOne(query string, args ...any) (*T, error)
	Insert(query string, args ...any) error
}

type QueryServiceImpl[T any] struct {
	DataAccessor database.DataAccessor
}

func (q *QueryServiceImpl[T]) Query(query string, args ...any) ([]T, error) {
	var results []T
	err := pgxscan.Select(context.Background(), q.DataAccessor.GetConnection(), &results, query, args...)
	return results, err
}

func (q *QueryServiceImpl[T]) QueryOne(query string, args ...any) (*T, error) {
	rows, err := q.Query(query, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

func (q *QueryServiceImpl[T]) Insert(query string, args ...any) error {
	_, err := q.DataAccessor.GetConnection().Exec(context.Background(), query, args...)
	return err
}
