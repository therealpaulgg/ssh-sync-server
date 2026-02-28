package models

import "github.com/google/uuid"

type KnownHost struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	HostPattern string    `json:"host_pattern" db:"host_pattern"`
	KeyType     string    `json:"key_type" db:"key_type"`
	KeyData     string    `json:"key_data" db:"key_data"`
	Marker      string    `json:"marker" db:"marker"`
}
