package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
)

// SshKeyChangeMock is a mock implementation of SshKeyChangeRepository
type SshKeyChangeMock struct {
	Changes []models.SshKeyChange
}

// CreateKeyChange records a new change to an SSH key
func (mock *SshKeyChangeMock) CreateKeyChange(change *models.SshKeyChange) (*models.SshKeyChange, error) {
	if change.ID == uuid.Nil {
		change.ID = uuid.New()
	}
	if change.ChangeTime.IsZero() {
		change.ChangeTime = time.Now()
	}
	
	mock.Changes = append(mock.Changes, *change)
	return change, nil
}

// CreateKeyChangeTx records a new change to an SSH key within a transaction
func (mock *SshKeyChangeMock) CreateKeyChangeTx(change *models.SshKeyChange, tx pgx.Tx) (*models.SshKeyChange, error) {
	return mock.CreateKeyChange(change)
}

// GetKeyChanges returns all changes for a specific SSH key
func (mock *SshKeyChangeMock) GetKeyChanges(sshKeyID uuid.UUID) ([]models.SshKeyChange, error) {
	var result []models.SshKeyChange
	
	for _, change := range mock.Changes {
		if change.SshKeyID == sshKeyID {
			result = append(result, change)
		}
	}
	
	return result, nil
}

// GetLatestKeyChangesForUser returns the most recent changes for each SSH key owned by a user
func (mock *SshKeyChangeMock) GetLatestKeyChangesForUser(userID uuid.UUID, since time.Time) ([]models.SshKeyChange, error) {
	var result []models.SshKeyChange
	keyMap := make(map[uuid.UUID]models.SshKeyChange)
	
	// Find the latest change for each key
	for _, change := range mock.Changes {
		if change.UserID == userID && change.ChangeTime.After(since) {
			existing, exists := keyMap[change.SshKeyID]
			if !exists || change.ChangeTime.After(existing.ChangeTime) {
				keyMap[change.SshKeyID] = change
			}
		}
	}
	
	// Convert map to slice
	for _, change := range keyMap {
		result = append(result, change)
	}
	
	return result, nil
}