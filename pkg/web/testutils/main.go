package testutils

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"net/http"

	"github.com/google/uuid"
	"github.com/therealpaulgg/ssh-sync-server/pkg/database/models"
	"github.com/therealpaulgg/ssh-sync-server/pkg/web/middleware/context_keys"
)

func GenerateUser() *models.User {
	return &models.User{
		ID:       uuid.New(),
		Username: "test",
	}
}

func GenerateMachine() *models.Machine {
	return &models.Machine{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Name:      "test",
		PublicKey: []byte("test"),
	}
}

func AddUserContext(req *http.Request, user *models.User) *http.Request {
	ctx := req.Context()
	ctx = context.WithValue(ctx, context_keys.UserContextKey, user)
	return req.Clone(ctx)
}

func AddMachineContext(req *http.Request, machine *models.Machine) *http.Request {
	ctx := req.Context()
	ctx = context.WithValue(ctx, context_keys.MachineContextKey, machine)
	return req.Clone(ctx)
}

// GenerateTestKeys generates a pair of ecdsa keys
func GenerateTestKeys() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	pubKey := &privKey.PublicKey
	return privKey, pubKey, nil
}

// EncodeToPem encodes keys into pem format
func EncodeToPem(privKey *ecdsa.PrivateKey, pubKey *ecdsa.PublicKey) ([]byte, []byte, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return nil, nil, err
	}

	privBytes, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return nil, nil, err
	}

	pubBytes = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	privBytes = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	return pubBytes, privBytes, nil
}
