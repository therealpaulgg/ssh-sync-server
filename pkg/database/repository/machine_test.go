package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database"
)

// MockTx is a mock implementation of pgx.Tx
type MockTx struct {
	queryExecuted string
	queryArgs     []interface{}
	commitCalled  bool
	rollbackCalled bool
	execError     error
	commitError   error
}

// Conn is required by pgx.Tx interface
func (m *MockTx) Conn() *pgx.Conn {
	return nil
}

func (m *MockTx) Begin(ctx context.Context) (pgx.Tx, error) {
	return nil, nil
}

func (m *MockTx) Commit(ctx context.Context) error {
	m.commitCalled = true
	return m.commitError
}

func (m *MockTx) Rollback(ctx context.Context) error {
	m.rollbackCalled = true
	return nil
}

func (m *MockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

func (m *MockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return nil
}

func (m *MockTx) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}

func (m *MockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}

func (m *MockTx) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	m.queryExecuted = sql
	m.queryArgs = arguments
	return pgconn.CommandTag{}, m.execError
}

func (m *MockTx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (m *MockTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return nil
}

// TestDeleteMachine verifies that the DeleteMachine function 
// no longer references SSH configs when deleting a machine
func TestDeleteMachine(t *testing.T) {
	// Arrange
	machineID := uuid.New()
	
	// Create a mock transaction that records what was executed
	mockTx := &MockTx{}
	
	// Setup mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	// Setup mock DataAccessor
	mockDataAccessor := database.NewMockDataAccessor(ctrl)
	
	// Mock connection 
	mockConn := &pgx.Conn{}
	
	// Set expectations
	mockDataAccessor.EXPECT().GetConnection().Return(mockConn).AnyTimes()
	
	// Create injector and provide mock
	injector := do.New()
	do.Provide(injector, func(i *do.Injector) (database.DataAccessor, error) {
		return mockDataAccessor, nil
	})
	
	// Create repository with injector
	repo := &MachineRepo{
		Injector: injector,
	}
	
	// Replace the BeginTx function for testing
	originalBeginTx := beginTxFunc
	defer func() { beginTxFunc = originalBeginTx }()
	
	// Mock the BeginTx function
	beginTxFunc = func(ctx context.Context, conn *pgx.Conn, txOptions pgx.TxOptions) (pgx.Tx, error) {
		return mockTx, nil
	}
	
	// Act
	err := repo.DeleteMachine(machineID)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "delete from machines where id = $1", mockTx.queryExecuted)
	assert.Equal(t, 1, len(mockTx.queryArgs))
	assert.Equal(t, machineID, mockTx.queryArgs[0])
	assert.True(t, mockTx.commitCalled)
	assert.False(t, mockTx.rollbackCalled)
}

// TestDeleteMachineExecError verifies that the transaction is rolled back
// when Exec returns an error
func TestDeleteMachineExecError(t *testing.T) {
	// Arrange
	machineID := uuid.New()
	execError := errors.New("exec error")
	
	// Create a mock transaction that returns an error
	mockTx := &MockTx{
		execError: execError,
	}
	
	// Setup mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	// Setup mock DataAccessor
	mockDataAccessor := database.NewMockDataAccessor(ctrl)
	
	// Mock connection 
	mockConn := &pgx.Conn{}
	
	// Set expectations
	mockDataAccessor.EXPECT().GetConnection().Return(mockConn).AnyTimes()
	
	// Create injector and provide mock
	injector := do.New()
	do.Provide(injector, func(i *do.Injector) (database.DataAccessor, error) {
		return mockDataAccessor, nil
	})
	
	// Create repository with injector
	repo := &MachineRepo{
		Injector: injector,
	}
	
	// Replace the BeginTx function for testing
	originalBeginTx := beginTxFunc
	defer func() { beginTxFunc = originalBeginTx }()
	
	// Mock the BeginTx function
	beginTxFunc = func(ctx context.Context, conn *pgx.Conn, txOptions pgx.TxOptions) (pgx.Tx, error) {
		return mockTx, nil
	}
	
	// Act
	err := repo.DeleteMachine(machineID)
	
	// Assert
	assert.Equal(t, execError, err)
	assert.Equal(t, "delete from machines where id = $1", mockTx.queryExecuted)
	assert.False(t, mockTx.commitCalled)
	assert.True(t, mockTx.rollbackCalled)
}

// TestDeleteMachineTxError verifies error handling when BeginTx fails
func TestDeleteMachineTxError(t *testing.T) {
	// Arrange
	machineID := uuid.New()
	txError := errors.New("transaction error")
	
	// Setup mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	// Setup mock DataAccessor
	mockDataAccessor := database.NewMockDataAccessor(ctrl)
	
	// Mock connection 
	mockConn := &pgx.Conn{}
	
	// Set expectations
	mockDataAccessor.EXPECT().GetConnection().Return(mockConn).AnyTimes()
	
	// Create injector and provide mock
	injector := do.New()
	do.Provide(injector, func(i *do.Injector) (database.DataAccessor, error) {
		return mockDataAccessor, nil
	})
	
	// Create repository with injector
	repo := &MachineRepo{
		Injector: injector,
	}
	
	// Replace the BeginTx function for testing
	originalBeginTx := beginTxFunc
	defer func() { beginTxFunc = originalBeginTx }()
	
	// Mock the BeginTx function to return an error
	beginTxFunc = func(ctx context.Context, conn *pgx.Conn, txOptions pgx.TxOptions) (pgx.Tx, error) {
		return nil, txError
	}
	
	// Act
	err := repo.DeleteMachine(machineID)
	
	// Assert
	assert.Equal(t, txError, err)
}

// Define a variable to hold the function so we can replace it in tests
var beginTxFunc = func(ctx context.Context, conn *pgx.Conn, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return conn.BeginTx(ctx, txOptions)
}
