package models

import (
	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID   `json:"id" db:"id"`
	Username   string      `json:"username" db:"username"`
	Keys       []SshKey    `json:"keys"`
	Config     []SshConfig `json:"config"`
	Machines   []Machine   `json:"machines"`
	KnownHosts []KnownHost `json:"known_hosts"`
}
