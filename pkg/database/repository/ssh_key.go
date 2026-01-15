package repository

import (
	"github.com/jackc/pgx/v5"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
)

type SshKeyRepository interface {
	CreateSshKey(sshKey *models.SshKey) (*models.SshKey, error)
	UpsertSshKey(sshKey *models.SshKey) (*models.SshKey, error)
	UpsertSshKeyTx(sshKey *models.SshKey, tx pgx.Tx) (*models.SshKey, error)
}

type SshKeyRepo struct {
	Injector *do.Injector
}

func (repo *SshKeyRepo) CreateSshKey(sshKey *models.SshKey) (*models.SshKey, error) {
	q := do.MustInvoke[query.QueryService[models.SshKey]](repo.Injector)
	key, err := q.QueryOne("INSERT INTO ssh_keys (user_id, filename, data, updated_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP) RETURNING *", sshKey.UserID, sshKey.Filename, sshKey.Data)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (repo *SshKeyRepo) UpsertSshKey(sshKey *models.SshKey) (*models.SshKey, error) {
	q := do.MustInvoke[query.QueryService[models.SshKey]](repo.Injector)
	key, err := q.QueryOne("INSERT INTO ssh_keys (user_id, filename, data, updated_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP) ON CONFLICT (user_id, filename) DO UPDATE SET data = $3, updated_at = CURRENT_TIMESTAMP RETURNING *", sshKey.UserID, sshKey.Filename, sshKey.Data)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (repo *SshKeyRepo) UpsertSshKeyTx(sshKey *models.SshKey, tx pgx.Tx) (*models.SshKey, error) {
	q := do.MustInvoke[query.QueryServiceTx[models.SshKey]](repo.Injector)
	key, err := q.QueryOne(tx, "INSERT INTO ssh_keys (user_id, filename, data, updated_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP) ON CONFLICT (user_id, filename) DO UPDATE SET data = $3, updated_at = CURRENT_TIMESTAMP RETURNING *", sshKey.UserID, sshKey.Filename, sshKey.Data)
	if err != nil {
		return nil, err
	}
	return key, nil
}
