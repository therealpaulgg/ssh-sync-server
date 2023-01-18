package models

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/do"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/query"
)

type SshKey struct {
	ID       uuid.UUID `json:"id" db:"id"`
	UserID   uuid.UUID `json:"user_id" db:"user_id"`
	Filename string    `json:"filename" db:"filename"`
	Data     []byte    `json:"data" db:"data"`
}

func (s *SshKey) CreateSshKey(i *do.Injector) error {
	q := do.MustInvoke[query.QueryService[SshKey]](i)
	key, err := q.QueryOne("INSERT INTO ssh_keys (user_id, filename, data) VALUES ($1, $2, $3) RETURNING *", s.UserID, s.Filename, s.Data)
	if err != nil {
		return err
	}
	s.ID = key.ID
	s.UserID = key.UserID
	s.Filename = key.Filename
	s.Data = key.Data
	return nil
}

func (s *SshKey) UpsertSshKey(i *do.Injector) error {
	q := do.MustInvoke[query.QueryService[SshKey]](i)
	key, err := q.QueryOne("INSERT INTO ssh_keys (user_id, filename, data) VALUES ($1, $2, $3) ON CONFLICT (user_id, filename) DO UPDATE SET data = $3 RETURNING *", s.UserID, s.Filename, s.Data)
	if err != nil {
		return err
	}
	s.ID = key.ID
	s.UserID = key.UserID
	s.Filename = key.Filename
	s.Data = key.Data
	return nil
}

func (s *SshKey) UpsertSshKeyTx(i *do.Injector, tx pgx.Tx) error {
	q := do.MustInvoke[query.QueryServiceTx[SshKey]](i)
	key, err := q.QueryOne(tx, "INSERT INTO ssh_keys (user_id, filename, data) VALUES ($1, $2, $3) ON CONFLICT (user_id, filename) DO UPDATE SET data = $3 RETURNING *", s.UserID, s.Filename, s.Data)
	if err != nil {
		return err
	}
	s.ID = key.ID
	s.UserID = key.UserID
	s.Filename = key.Filename
	s.Data = key.Data
	return nil
}
