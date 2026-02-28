package models

import (
	"time"

	"github.com/google/uuid"
)

type MasterKeyRotation struct {
	ID                 uuid.UUID `db:"id"`
	MachineID          uuid.UUID `db:"machine_id"`
	EncryptedMasterKey []byte    `db:"encrypted_master_key"`
	CreatedAt          time.Time `db:"created_at"`
}
