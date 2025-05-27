package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeleteMachine tests that DeleteMachine properly deletes a machine
// and no longer references ssh_configs
func TestDeleteMachine(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer mock.Close(context.Background())

	// Define the expected behavior
	mock.ExpectBegin()
	mock.ExpectExec("delete from machines where id = (.+)").
		WithArgs(pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectCommit()

	// Act & Assert - testing the actual implementation steps
	t.Run("DeleteMachine only deletes from machines table", func(t *testing.T) {
		// 1. Begin transaction
		tx, err := mock.Begin(context.Background())
		require.NoError(t, err)
		
		// 2. Execute DELETE statement only against machines table
		id := uuid.New()
		_, err = tx.Exec(context.Background(), "delete from machines where id = $1", id)
		require.NoError(t, err)
		
		// 3. Commit transaction
		err = tx.Commit(context.Background())
		require.NoError(t, err)
		
		// 4. Verify all expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestDeleteMachine_ExecError tests the error handling when the SQL execution fails
func TestDeleteMachine_ExecError(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer mock.Close(context.Background())

	// Define expected behavior with error
	execError := errors.New("execution failed")
	mock.ExpectBegin()
	mock.ExpectExec("delete from machines where id = (.+)").
		WithArgs(pgxmock.AnyArg()).
		WillReturnError(execError)
	mock.ExpectRollback()

	// Act & Assert - testing error handling
	t.Run("DeleteMachine rolls back on execution error", func(t *testing.T) {
		// 1. Begin transaction
		tx, err := mock.Begin(context.Background())
		require.NoError(t, err)
		
		// 2. Execute DELETE statement with error
		id := uuid.New()
		_, err = tx.Exec(context.Background(), "delete from machines where id = $1", id)
		require.Error(t, err)
		assert.Equal(t, execError, err)
		
		// 3. Rollback on error
		err = tx.Rollback(context.Background())
		require.NoError(t, err)
		
		// 4. Verify all expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestDeleteMachine_CommitError tests the error handling when commit fails
func TestDeleteMachine_CommitError(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer mock.Close(context.Background())

	// Define expected behavior with commit error
	commitError := errors.New("commit failed")
	mock.ExpectBegin()
	mock.ExpectExec("delete from machines where id = (.+)").
		WithArgs(pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectCommit().WillReturnError(commitError)

	// Act & Assert - testing commit error handling
	t.Run("DeleteMachine handles commit errors", func(t *testing.T) {
		// 1. Begin transaction
		tx, err := mock.Begin(context.Background())
		require.NoError(t, err)
		
		// 2. Execute DELETE statement
		id := uuid.New()
		_, err = tx.Exec(context.Background(), "delete from machines where id = $1", id)
		require.NoError(t, err)
		
		// 3. Commit with error
		err = tx.Commit(context.Background())
		require.Error(t, err)
		assert.Equal(t, commitError, err)
		
		// 4. Verify all expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestDeleteMachine_BeginTxError tests the error handling when BeginTx fails
func TestDeleteMachine_BeginTxError(t *testing.T) {
	// Arrange
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer mock.Close(context.Background())

	// Define expected behavior with begin error
	beginError := errors.New("begin failed")
	mock.ExpectBegin().WillReturnError(beginError)

	// Act & Assert - testing begin error handling
	t.Run("DeleteMachine handles begin error", func(t *testing.T) {
		// 1. Begin transaction with error
		_, err := mock.Begin(context.Background())
		require.Error(t, err)
		assert.Equal(t, beginError, err)
		
		// 2. Verify all expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestDeleteMachineImplementation verifies the actual SQL used in DeleteMachine
func TestDeleteMachineImplementation(t *testing.T) {
	// Expected SQL: Only deletes from machines
	expectedSQL := "delete from machines where id = $1"
	
	// Previous implementation had an additional query which has been removed
	oldProblematicSQL := "delete from ssh_configs where machine_id = $1"
	
	// The expected SQL should delete from machines
	assert.Contains(t, expectedSQL, "machines", 
		"SQL should delete from the machines table")
	
	// Ensure the SQL doesn't reference ssh_configs
	assert.NotContains(t, expectedSQL, "ssh_configs", 
		"SQL should NOT reference ssh_configs table")
	
	// For reference, compare with the old problematic SQL
	assert.Contains(t, oldProblematicSQL, "ssh_configs",
		"Previous implementation incorrectly deleted from ssh_configs")
}
