package query

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database"
)

type TransactionService interface {
	StartTx(options pgx.TxOptions) (pgx.Tx, error)
	Commit(tx pgx.Tx) error
	Rollback(tx pgx.Tx) error
}

type TransactionServiceImpl struct {
	DataAccessor database.DataAccessor
}

func (q *TransactionServiceImpl) StartTx(options pgx.TxOptions) (pgx.Tx, error) {
	var err error
	tx, err := q.DataAccessor.GetConnection().BeginTx(context.Background(), options)
	return tx, err
}

func (q *TransactionServiceImpl) Commit(tx pgx.Tx) error {
	return tx.Commit(context.Background())
}

func (q *TransactionServiceImpl) Rollback(tx pgx.Tx) error {
	return tx.Rollback(context.Background())
}

type QueryServiceTx[T any] interface {
	Query(tx pgx.Tx, query string, args ...interface{}) ([]T, error)
	QueryOne(tx pgx.Tx, query string, args ...interface{}) (*T, error)
	Insert(tx pgx.Tx, query string, args ...interface{}) error
}

type QueryServiceTxImpl[T any] struct {
	DataAccessor database.DataAccessor
}

func (q *QueryServiceTxImpl[T]) Query(tx pgx.Tx, query string, args ...interface{}) ([]T, error) {
	var results []T
	err := pgxscan.Select(context.Background(), tx, &results, query, args...)
	return results, err
}

func (q *QueryServiceTxImpl[T]) QueryOne(tx pgx.Tx, query string, args ...interface{}) (*T, error) {
	rows, err := q.Query(tx, query, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

func (q *QueryServiceTxImpl[T]) Insert(tx pgx.Tx, query string, args ...interface{}) error {
	_, err := tx.Exec(context.Background(), query, args...)
	return err
}
