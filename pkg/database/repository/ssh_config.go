package repository

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
)

type SshConfigRepository interface {
	GetSshConfig(machineID uuid.UUID, userID uuid.UUID) (*models.SshConfig, error)
	UpsertSshConfig(config *models.SshConfig) (*models.SshConfig, error)
	UpsertSshConfigTx(config *models.SshConfig, tx pgx.Tx) (*models.SshConfig, error)
}

type SshConfigRepo struct {
	Injector *do.Injector
}

func (repo *SshConfigRepo) GetSshConfig(machineID uuid.UUID, userID uuid.UUID) (*models.SshConfig, error) {
	q := do.MustInvoke[query.QueryService[models.SshConfig]](repo.Injector)
	sshConfig, err := q.QueryOne("select * from ssh_configs where machine_id = $1 and user_id = $2", machineID, userID)
	if err != nil {
		return nil, err
	}
	if sshConfig == nil {
		return nil, sql.ErrNoRows
	}
	return sshConfig, nil
}

func (repo *SshConfigRepo) UpsertSshConfig(config *models.SshConfig) (*models.SshConfig, error) {
	q := do.MustInvoke[query.QueryService[models.SshConfig]](repo.Injector)
	sshConfig, err := q.QueryOne("insert into ssh_configs (user_id, machine_id, host, values, identity_files) values ($1, $2, $3, $4, $5) on conflict (user_id, machine_id, host) do update set host = $3, values = $4, identity_files = $5 returning *", config.UserID, config.MachineID, config.Host, config.Values, config.IdentityFiles)
	if err != nil {
		return nil, err
	}
	if sshConfig == nil {
		return nil, sql.ErrNoRows
	}
	return sshConfig, nil
}

func (repo *SshConfigRepo) UpsertSshConfigTx(config *models.SshConfig, tx pgx.Tx) (*models.SshConfig, error) {
	q := do.MustInvoke[query.QueryServiceTx[models.SshConfig]](repo.Injector)
	sshConfig, err := q.QueryOne(tx, "insert into ssh_configs (user_id, machine_id, host, values, identity_files) values ($1, $2, $3, $4, $5) on conflict (user_id, machine_id, host) do update set host = $3, values = $4, identity_files = $5 returning *", config.UserID, config.MachineID, config.Host, config.Values, config.IdentityFiles)
	if err != nil {
		return nil, err
	}
	if sshConfig == nil {
		return nil, sql.ErrNoRows
	}
	return sshConfig, nil
}
