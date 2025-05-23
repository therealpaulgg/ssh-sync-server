package repository

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
)

type SshKeyRepository interface {
	CreateSshKey(sshKey *models.SshKey) (*models.SshKey, error)
	UpsertSshKey(sshKey *models.SshKey) (*models.SshKey, error)
	UpsertSshKeyTx(sshKey *models.SshKey, tx pgx.Tx) (*models.SshKey, error)
	// Methods with change tracking
	CreateSshKeyWithChange(sshKey *models.SshKey) (*models.SshKey, error)
	UpsertSshKeyWithChange(sshKey *models.SshKey) (*models.SshKey, error)
	UpsertSshKeyWithChangeTx(sshKey *models.SshKey, tx pgx.Tx) (*models.SshKey, error)
	// Get a key by user ID and filename
	GetSshKeyByFilename(userID uuid.UUID, filename string) (*models.SshKey, error)
}

type SshKeyRepo struct {
	Injector *do.Injector
}

func (repo *SshKeyRepo) CreateSshKey(sshKey *models.SshKey) (*models.SshKey, error) {
	q := do.MustInvoke[query.QueryService[models.SshKey]](repo.Injector)
	key, err := q.QueryOne("INSERT INTO ssh_keys (user_id, filename, data) VALUES ($1, $2, $3) RETURNING *", sshKey.UserID, sshKey.Filename, sshKey.Data)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (repo *SshKeyRepo) UpsertSshKey(sshKey *models.SshKey) (*models.SshKey, error) {
	q := do.MustInvoke[query.QueryService[models.SshKey]](repo.Injector)
	key, err := q.QueryOne("INSERT INTO ssh_keys (user_id, filename, data) VALUES ($1, $2, $3) ON CONFLICT (user_id, filename) DO UPDATE SET data = $3 RETURNING *", sshKey.UserID, sshKey.Filename, sshKey.Data)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (repo *SshKeyRepo) UpsertSshKeyTx(sshKey *models.SshKey, tx pgx.Tx) (*models.SshKey, error) {
	q := do.MustInvoke[query.QueryServiceTx[models.SshKey]](repo.Injector)
	key, err := q.QueryOne(tx, "INSERT INTO ssh_keys (user_id, filename, data) VALUES ($1, $2, $3) ON CONFLICT (user_id, filename) DO UPDATE SET data = $3 RETURNING *", sshKey.UserID, sshKey.Filename, sshKey.Data)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (repo *SshKeyRepo) GetSshKeyByFilename(userID uuid.UUID, filename string) (*models.SshKey, error) {
	q := do.MustInvoke[query.QueryService[models.SshKey]](repo.Injector)
	key, err := q.QueryOne("SELECT * FROM ssh_keys WHERE user_id = $1 AND filename = $2", userID, filename)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (repo *SshKeyRepo) CreateSshKeyWithChange(sshKey *models.SshKey) (*models.SshKey, error) {
	// Start a transaction
	txService := do.MustInvoke[query.TransactionService](repo.Injector)
	tx, err := txService.StartTx(pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	
	// Defer rollback in case of error
	defer func() {
		if err != nil {
			_ = txService.Rollback(tx)
		}
	}()
	
	// Create the SSH key
	key, err := repo.UpsertSshKeyTx(sshKey, tx)
	if err != nil {
		return nil, err
	}
	
	// Record the change
	changeRepo := &SshKeyChangeRepo{Injector: repo.Injector}
	change := &models.SshKeyChange{
		SshKeyID:   key.ID,
		UserID:     key.UserID,
		ChangeType: models.Created,
		Filename:   key.Filename,
		NewData:    key.Data,
	}
	
	_, err = changeRepo.CreateKeyChangeTx(change, tx)
	if err != nil {
		return nil, err
	}
	
	// Commit the transaction
	err = txService.Commit(tx)
	if err != nil {
		return nil, err
	}
	
	return key, nil
}

func (repo *SshKeyRepo) UpsertSshKeyWithChange(sshKey *models.SshKey) (*models.SshKey, error) {
	// Start a transaction
	txService := do.MustInvoke[query.TransactionService](repo.Injector)
	tx, err := txService.StartTx(pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	
	// Defer rollback in case of error
	defer func() {
		if err != nil {
			_ = txService.Rollback(tx)
		}
	}()
	
	// Get the existing key if it exists
	var existingKey *models.SshKey
	var changeType models.ChangeType
	
	existingKey, err = repo.GetSshKeyByFilename(sshKey.UserID, sshKey.Filename)
	if err != nil {
		// If error is not "no rows", return the error
		// Otherwise, continue as it's a new key
		changeType = models.Created
	} else if existingKey != nil {
		changeType = models.Updated
	} else {
		changeType = models.Created
	}
	
	// Upsert the SSH key
	key, err := repo.UpsertSshKeyTx(sshKey, tx)
	if err != nil {
		return nil, err
	}
	
	// Record the change
	changeRepo := &SshKeyChangeRepo{Injector: repo.Injector}
	change := &models.SshKeyChange{
		SshKeyID:   key.ID,
		UserID:     key.UserID,
		ChangeType: changeType,
		Filename:   key.Filename,
		NewData:    key.Data,
	}
	
	if existingKey != nil {
		change.PreviousData = existingKey.Data
	}
	
	_, err = changeRepo.CreateKeyChangeTx(change, tx)
	if err != nil {
		return nil, err
	}
	
	// Commit the transaction
	err = txService.Commit(tx)
	if err != nil {
		return nil, err
	}
	
	return key, nil
}

func (repo *SshKeyRepo) UpsertSshKeyWithChangeTx(sshKey *models.SshKey, tx pgx.Tx) (*models.SshKey, error) {
	// Get the existing key if it exists
	var existingKey *models.SshKey
	var changeType models.ChangeType
	
	existingKey, err := repo.GetSshKeyByFilename(sshKey.UserID, sshKey.Filename)
	if err != nil {
		// If error is not "no rows", return the error
		// Otherwise, continue as it's a new key
		changeType = models.Created
	} else if existingKey != nil {
		changeType = models.Updated
	} else {
		changeType = models.Created
	}
	
	// Upsert the SSH key
	key, err := repo.UpsertSshKeyTx(sshKey, tx)
	if err != nil {
		return nil, err
	}
	
	// Record the change
	changeRepo := &SshKeyChangeRepo{Injector: repo.Injector}
	change := &models.SshKeyChange{
		SshKeyID:   key.ID,
		UserID:     key.UserID,
		ChangeType: changeType,
		Filename:   key.Filename,
		NewData:    key.Data,
	}
	
	if existingKey != nil {
		change.PreviousData = existingKey.Data
	}
	
	_, err = changeRepo.CreateKeyChangeTx(change, tx)
	if err != nil {
		return nil, err
	}
	
	return key, nil
}
