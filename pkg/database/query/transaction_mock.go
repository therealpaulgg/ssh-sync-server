package query

import (
	"github.com/jackc/pgx/v5"
)

// TransactionMock implements the TransactionService interface for testing
type TransactionMock struct {
	StartTxFunc  func(pgx.TxOptions) (pgx.Tx, error)
	CommitFunc   func(pgx.Tx) error
	RollbackFunc func(pgx.Tx) error
}

// StartTx implements TransactionService
func (m *TransactionMock) StartTx(opts pgx.TxOptions) (pgx.Tx, error) {
	if m.StartTxFunc != nil {
		return m.StartTxFunc(opts)
	}
	return nil, nil
}

// Commit implements TransactionService
func (m *TransactionMock) Commit(tx pgx.Tx) error {
	if m.CommitFunc != nil {
		return m.CommitFunc(tx)
	}
	return nil
}

// Rollback implements TransactionService
func (m *TransactionMock) Rollback(tx pgx.Tx) error {
	if m.RollbackFunc != nil {
		return m.RollbackFunc(tx)
	}
	return nil
}