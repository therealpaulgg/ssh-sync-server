package models

import "github.com/google/uuid"

type SshConfig struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	MachineID uuid.UUID `json:"machine_id" db:"machine_id"`
	Host      string    `json:"host" db:"host"`
	Values    string    `json:"values" db:"values"`
}
