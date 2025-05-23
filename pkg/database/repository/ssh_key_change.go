package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
)

// SshKeyChangeRepository defines the interface for operations on SSH key changes
type SshKeyChangeRepository interface {
	// CreateKeyChange records a new change to an SSH key
	CreateKeyChange(change *models.SshKeyChange) (*models.SshKeyChange, error)
	// CreateKeyChangeTx records a new change to an SSH key within a transaction
	CreateKeyChangeTx(change *models.SshKeyChange, tx pgx.Tx) (*models.SshKeyChange, error)
	// GetKeyChanges returns all changes for a specific SSH key
	GetKeyChanges(sshKeyID uuid.UUID) ([]models.SshKeyChange, error)
	// GetLatestKeyChangesForUser returns the most recent changes for each SSH key owned by a user
	GetLatestKeyChangesForUser(userID uuid.UUID, since time.Time) ([]models.SshKeyChange, error)
}

// SshKeyChangeRepo implements the SshKeyChangeRepository interface
type SshKeyChangeRepo struct {
	Injector *do.Injector
}

// CreateKeyChange records a new change to an SSH key
func (repo *SshKeyChangeRepo) CreateKeyChange(change *models.SshKeyChange) (*models.SshKeyChange, error) {
	q := do.MustInvoke[query.QueryService[models.SshKeyChange]](repo.Injector)
	
	// Set the ID and timestamp if not already set
	if change.ID == uuid.Nil {
		change.ID = uuid.New()
	}
	if change.ChangeTime.IsZero() {
		change.ChangeTime = time.Now()
	}
	
	result, err := q.QueryOne(
		"INSERT INTO ssh_key_changes (id, ssh_key_id, user_id, change_type, filename, previous_data, new_data, change_time) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *",
		change.ID, change.SshKeyID, change.UserID, change.ChangeType, 
		change.Filename, change.PreviousData, change.NewData, change.ChangeTime)
	
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

// CreateKeyChangeTx records a new change to an SSH key within a transaction
func (repo *SshKeyChangeRepo) CreateKeyChangeTx(change *models.SshKeyChange, tx pgx.Tx) (*models.SshKeyChange, error) {
	q := do.MustInvoke[query.QueryServiceTx[models.SshKeyChange]](repo.Injector)
	
	// Set the ID and timestamp if not already set
	if change.ID == uuid.Nil {
		change.ID = uuid.New()
	}
	if change.ChangeTime.IsZero() {
		change.ChangeTime = time.Now()
	}
	
	result, err := q.QueryOne(
		tx,
		"INSERT INTO ssh_key_changes (id, ssh_key_id, user_id, change_type, filename, previous_data, new_data, change_time) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *",
		change.ID, change.SshKeyID, change.UserID, change.ChangeType, 
		change.Filename, change.PreviousData, change.NewData, change.ChangeTime)
	
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

// GetKeyChanges returns all changes for a specific SSH key
func (repo *SshKeyChangeRepo) GetKeyChanges(sshKeyID uuid.UUID) ([]models.SshKeyChange, error) {
	q := do.MustInvoke[query.QueryService[models.SshKeyChange]](repo.Injector)
	
	results, err := q.Query(
		"SELECT * FROM ssh_key_changes WHERE ssh_key_id = $1 ORDER BY change_time DESC",
		sshKeyID)
	
	if err != nil {
		return nil, err
	}
	
	return results, nil
}

// GetLatestKeyChangesForUser returns the most recent changes for each SSH key owned by a user
func (repo *SshKeyChangeRepo) GetLatestKeyChangesForUser(userID uuid.UUID, since time.Time) ([]models.SshKeyChange, error) {
	q := do.MustInvoke[query.QueryService[models.SshKeyChange]](repo.Injector)
	
	results, err := q.Query(
		`SELECT DISTINCT ON (ssh_key_id) * 
		FROM ssh_key_changes 
		WHERE user_id = $1 AND change_time > $2
		ORDER BY ssh_key_id, change_time DESC`,
		userID, since)
	
	if err != nil {
		return nil, err
	}
	
	return results, nil
}