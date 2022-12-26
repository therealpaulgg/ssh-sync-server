package models

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
)

type SshConfig struct {
	ID           uuid.UUID         `json:"id" db:"id"`
	UserID       uuid.UUID         `json:"user_id" db:"user_id"`
	MachineID    uuid.UUID         `json:"machine_id" db:"machine_id"`
	Host         string            `json:"host" db:"host"`
	Values       map[string]string `json:"values" db:"values"`
	IdentityFile string            `json:"identity_file" db:"identity_file"`
}

func (s *SshConfig) GetSshConfig(i *do.Injector) error {
	q := do.MustInvoke[query.QueryService[SshConfig]](i)
	sshConfig, err := q.QueryOne("select * from ssh_configs where machine_id = $1 and user_id = $2", s.MachineID, s.UserID)
	if err != nil {
		return err
	}
	if sshConfig == nil {
		return sql.ErrNoRows
	}
	s.ID = sshConfig.ID
	s.UserID = sshConfig.UserID
	s.MachineID = sshConfig.MachineID
	s.Host = sshConfig.Host
	s.Values = sshConfig.Values
	s.IdentityFile = sshConfig.IdentityFile
	return nil
}

func (s *SshConfig) UpsertSshConfig(i *do.Injector) error {
	q := do.MustInvoke[query.QueryService[SshConfig]](i)
	sshConfig, err := q.QueryOne("insert into ssh_configs (user_id, machine_id, host, values, identity_file) values ($1, $2, $3, $4, $5) on conflict (user_id, machine_id, host) do update set host = $3, values = $4, identity_file = $5 returning *", s.UserID, s.MachineID, s.Host, s.Values, s.IdentityFile)
	if err != nil {
		return err
	}
	if sshConfig == nil {
		return sql.ErrNoRows
	}
	s.ID = sshConfig.ID
	s.UserID = sshConfig.UserID
	s.MachineID = sshConfig.MachineID
	s.Host = sshConfig.Host
	s.Values = sshConfig.Values
	s.IdentityFile = sshConfig.IdentityFile
	return nil
}
