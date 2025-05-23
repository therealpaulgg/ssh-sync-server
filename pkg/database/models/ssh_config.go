package models

import (
	"github.com/google/uuid"
)

type SshConfig struct {
	ID            uuid.UUID           `json:"id" db:"id"`
	UserID        uuid.UUID           `json:"user_id" db:"user_id"`
	Host          string              `json:"host" db:"host"`
	Values        map[string][]string `json:"values" db:"values"`
	IdentityFiles []string            `json:"identity_files" db:"identity_files"`
	KnownHosts    []byte              `json:"known_hosts" db:"known_hosts"`
}
