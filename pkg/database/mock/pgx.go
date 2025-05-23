package mock

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// MockPgxTx is a mock implementation of pgx.Tx for testing
type MockPgxTx struct {
	QueryExecuted string
	QueryArgs     []interface{}
	CommitCalled  bool
	RollbackCalled bool
	ExecError     error
	CommitError   error
}

// Conn is required by pgx.Tx interface
func (m *MockPgxTx) Conn() *pgx.Conn {
	return nil
}

func (m *MockPgxTx) Begin(ctx context.Context) (pgx.Tx, error) {
	return nil, nil
}

func (m *MockPgxTx) Commit(ctx context.Context) error {
	m.CommitCalled = true
	return m.CommitError
}

func (m *MockPgxTx) Rollback(ctx context.Context) error {
	m.RollbackCalled = true
	return nil
}

func (m *MockPgxTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

func (m *MockPgxTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return nil
}

func (m *MockPgxTx) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}

func (m *MockPgxTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}

func (m *MockPgxTx) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	m.QueryExecuted = sql
	m.QueryArgs = arguments
	return pgconn.CommandTag{}, m.ExecError
}

func (m *MockPgxTx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (m *MockPgxTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return nil
}

// MockPgxConn is a mock implementation of pgx.Conn that returns predefined transaction objects
type MockPgxConn struct {
	pgx.Conn // Embed the interface to satisfy it
	Tx       pgx.Tx
	TxError  error
}

// BeginTx returns the predefined transaction or error
func (m *MockPgxConn) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	if m.TxError != nil {
		return nil, m.TxError
	}
	return m.Tx, nil
}