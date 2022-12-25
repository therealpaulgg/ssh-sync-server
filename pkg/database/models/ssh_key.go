package models

import (
	"github.com/google/uuid"
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
	key, err := q.QueryOne("INSERT INTO ssh_keys (user_id, filename, data) VALUES ($1, $2, $3) RETURNING id", s.UserID, s.Filename, s.Data)
	if err != nil {
		return err
	}
	s.ID = key.ID
	s.UserID = key.UserID
	s.Filename = key.Filename
	s.Data = key.Data
	return nil
}