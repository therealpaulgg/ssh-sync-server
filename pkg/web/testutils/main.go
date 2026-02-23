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

// GenerateTestKeys generates a pair of ECDSA P-256 keys.
func GenerateTestKeys() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	pubKey := &privKey.PublicKey
	return privKey, pubKey, nil
}

// EncodeToPem encodes ECDSA keys into PEM format.
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

// GenerateTestEncapsulationKeyPEM generates a fake ML-KEM-768 encapsulation key PEM for testing.
// The key data is random bytes of the correct size (1184 bytes).
func GenerateTestEncapsulationKeyPEM() ([]byte, error) {
	keyBytes := make([]byte, 1184)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "MLKEM768 ENCAPSULATION KEY",
		Bytes: keyBytes,
	}), nil
}

// GenerateHybridKeyPEM generates a two-block PEM with EC public key + ML-KEM-768 encapsulation key.
func GenerateHybridKeyPEM() ([]byte, []byte, []byte, error) {
	priv, pub, err := GenerateTestKeys()
	if err != nil {
		return nil, nil, nil, err
	}
	pubPEM, _, err := EncodeToPem(priv, pub)
	if err != nil {
		return nil, nil, nil, err
	}
	encapPEM, err := GenerateTestEncapsulationKeyPEM()
	if err != nil {
		return nil, nil, nil, err
	}
	// Concatenate both PEM blocks
	combined := append(pubPEM, encapPEM...)
	return combined, pubPEM, encapPEM, nil
}
