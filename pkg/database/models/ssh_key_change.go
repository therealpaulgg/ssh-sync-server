package models

import (
	"time"

	"github.com/google/uuid"
)

// ChangeType represents the type of change made to an SSH key
type ChangeType string

const (
	// Created indicates a new SSH key was created
	Created ChangeType = "created"
	// Updated indicates an existing SSH key was updated
	Updated ChangeType = "updated"
	// Deleted indicates an SSH key was deleted
	Deleted ChangeType = "deleted"
)

// SshKeyChange represents a change to an SSH key in the database
type SshKeyChange struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	SshKeyID      uuid.UUID  `json:"ssh_key_id" db:"ssh_key_id"`
	UserID        uuid.UUID  `json:"user_id" db:"user_id"`
	ChangeType    ChangeType `json:"change_type" db:"change_type"`
	Filename      string     `json:"filename" db:"filename"`
	PreviousData  []byte     `json:"previous_data,omitempty" db:"previous_data"`
	NewData       []byte     `json:"new_data,omitempty" db:"new_data"`
	ChangeTime    time.Time  `json:"change_time" db:"change_time"`
}