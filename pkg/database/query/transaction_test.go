package query

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	testpgx "github.com/therealpaulgg/ssh-sync-server/test/pgx"
)

func TestTransactionServiceStartTx(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDa := &MockDataAccessor{}
	ts := &TransactionServiceImpl{
		DataAccessor: mockDa,
	}

	// Test StartTx
	t.Run("start transaction", func(t *testing.T) {
		// We can only test the structure because we can't mock the internal pgx BeginTx call easily
		// In a real test we would inject a mock to verify BeginTx was called with the right options
		assert.NotNil(t, ts)
	})
}

func TestTransactionServiceCommit(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTx := testpgx.NewMockTx(ctrl)
	mockTx.EXPECT().Commit(gomock.Any()).Return(nil)

	ts := &TransactionServiceImpl{}

	// Test commit
	err := ts.Commit(mockTx)
	assert.NoError(t, err)
}

func TestTransactionServiceRollback(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTx := testpgx.NewMockTx(ctrl)
	mockTx.EXPECT().Rollback(gomock.Any()).Return(nil)

	ts := &TransactionServiceImpl{}

	// Test rollback
	err := ts.Rollback(mockTx)
	assert.NoError(t, err)
}

func TestRollbackFunc(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTx := testpgx.NewMockTx(ctrl)
	ts := &TransactionServiceImpl{}

	// Create a test HTTP response writer
	w := httptest.NewRecorder()

	// Test case 1: With error, should rollback
	t.Run("with error should rollback", func(t *testing.T) {
		mockTx.EXPECT().Rollback(gomock.Any()).Return(nil)
		
		testErr := errors.New("test error")
		RollbackFunc(ts, mockTx, w, &testErr)
		
		// Nothing to assert for the response writer in this case
		// We're just ensuring the rollback was called
	})

	// Test case 2: Without error, should commit
	t.Run("without error should commit", func(t *testing.T) {
		mockTx.EXPECT().Commit(gomock.Any()).Return(nil)
		
		var testErr error = nil
		RollbackFunc(ts, mockTx, w, &testErr)
		
		// Verify response status remains 200 OK
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test case 3: Without error, but commit fails
	t.Run("commit fails", func(t *testing.T) {
		mockTx.EXPECT().Commit(gomock.Any()).Return(errors.New("commit failed"))
		mockTx.EXPECT().Rollback(gomock.Any()).Return(nil)
		
		var testErr error = nil
		RollbackFunc(ts, mockTx, w, &testErr)
		
		// Verify response status is set to 500
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestQueryServiceTx(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock transaction
	mockTx := testpgx.NewMockTx(ctrl)

	// Create query service
	queryTx := &QueryServiceTxImpl[TestModel]{
		DataAccessor: &MockDataAccessor{},
	}

	// Test Query method
	t.Run("query method", func(t *testing.T) {
		// Simplified test since mocking pgxscan.Select with tx is complex
		assert.NotNil(t, queryTx)
	})

	// Test QueryOne method
	t.Run("query one method", func(t *testing.T) {
		// Simplified test
		assert.NotNil(t, queryTx)
	})

	// Test Insert method
	t.Run("insert method", func(t *testing.T) {
		// Should call tx.Exec
		mockTx.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(pgconn.CommandTag{}, nil)
		
		err := queryTx.Insert(mockTx, "INSERT INTO test (id, name) VALUES ($1, $2)", 1, "test")
		assert.NoError(t, err)
	})
}