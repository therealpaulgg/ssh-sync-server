package models

import "github.com/google/uuid"

type Machine struct {
	ID     uuid.UUID `json:"id" db:"id"`
	UserID uuid.UUID `json:"user_id" db:"user_id"`
	Name   string    `json:"name" db:"name"`
}
