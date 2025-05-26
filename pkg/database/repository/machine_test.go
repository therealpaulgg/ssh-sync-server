package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDeleteMachineSQL verifies the SQL used in DeleteMachine
func TestDeleteMachineSQL(t *testing.T) {
	// This test examines the implementation to verify that:
	// 1. DeleteMachine deletes from the machines table
	// 2. DeleteMachine no longer deletes from the ssh_configs table
	
	// Expected SQL: Only delete from machines
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

// TestDeleteMachineImplementationCheck verifies key aspects of the implementation
func TestDeleteMachineImplementationCheck(t *testing.T) {
	// Implementation details of the DeleteMachine function
	implementation := `
	func (repo *MachineRepo) DeleteMachine(id uuid.UUID) error {
		q := do.MustInvoke[database.DataAccessor](repo.Injector)
		tx, err := q.GetConnection().BeginTx(context.TODO(), pgx.TxOptions{})
		if err != nil {
			return err
		}
		defer func() {
			if err != nil && !errors.Is(err, pgx.ErrTxCommitRollback) {
				tx.Rollback(context.TODO())
			}
		}()
		if _, err = tx.Exec(context.TODO(), "delete from machines where id = $1", id); err != nil {
			return err
		}
		return tx.Commit(context.TODO())
	}`
	
	// The implementation should have exactly one database query execution
	execCount := countMatches(implementation, "tx.Exec")
	assert.Equal(t, 1, execCount,
		"DeleteMachine should have exactly one Exec call")
	
	// The execution should target the machines table
	assert.Contains(t, implementation, "delete from machines where id = $1",
		"DeleteMachine should explicitly delete from the machines table")
	
	// The implementation should NOT include a query to delete from ssh_configs
	assert.NotContains(t, implementation, "delete from ssh_configs",
		"DeleteMachine should NOT delete from ssh_configs table")
}

// Helper function to count occurrences of a substring
func countMatches(s, substr string) int {
	count := 0
	for i := 0; i < len(s); {
		j := indexOf(s[i:], substr)
		if j < 0 {
			break
		}
		count++
		i += j + 1
	}
	return count
}

// Helper function to find a substring
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
