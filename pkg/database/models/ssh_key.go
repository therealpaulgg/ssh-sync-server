package models

import (
	"time"

	"github.com/google/uuid"
)

type SshKey struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	Filename  string     `json:"filename" db:"filename"`
	Data      []byte     `json:"data" db:"data"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
}
