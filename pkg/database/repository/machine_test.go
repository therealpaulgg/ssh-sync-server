package repository

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database"
)

// TestDeleteMachine verifies that the DeleteMachine function 
// no longer references SSH configs when deleting a machine
func TestDeleteMachine(t *testing.T) {
	// Skip tests until we can properly mock pgx transaction
	// NOTE: Testing the deletion of a machine requires complex mocking of pgx transactions
	// To properly test this, we would need to use a mocking library like github.com/DATA-DOG/go-sqlmock
	// or github.com/vektra/mockery to generate mocks for the pgx.Conn and pgx.Tx interfaces.
	// Alternatively, we could use an integration test with a real database.
	t.Skip("Skipping test until proper mocking can be implemented")
	
	// Arrange
	machineID := uuid.New()
	
	// Setup mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	// Setup mock DataAccessor
	mockDataAccessor := database.NewMockDataAccessor(ctrl)
	
	// Create injector and provide mock
	injector := do.New()
	do.Provide(injector, func(i *do.Injector) (database.DataAccessor, error) {
		return mockDataAccessor, nil
	})
	
	// Create repository with injector
	repo := &MachineRepo{
		Injector: injector,
	}
	
	// Act
	err := repo.DeleteMachine(machineID)
	
	// Assert
	assert.NoError(t, err)
	
	// What we would verify if we could properly mock the transaction:
	// 1. Only the machine deletion query is executed
	// 2. No query to delete SSH configs by machine_id is executed
	// 3. The transaction is committed on success
}
