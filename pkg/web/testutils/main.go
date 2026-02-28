package testutils

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"time"

	"filippo.io/mldsa"
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

// GenerateMLDSATestKeys generates an ML-DSA keypair.
func GenerateMLDSATestKeys() (*mldsa.PublicKey, *mldsa.PrivateKey, error) {
	priv, err := mldsa.GenerateKey(mldsa.MLDSA65())
	if err != nil {
		return nil, nil, err
	}
	return priv.PublicKey(), priv, nil
}

// EncodeMLDSAToPem PEM-encodes an ML-DSA public key.
func EncodeMLDSAToPem(pub *mldsa.PublicKey) ([]byte, error) {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "ML-DSA PUBLIC KEY",
		Bytes: pub.Bytes(),
	}), nil
}

// GenerateMLDSATestToken creates and signs a JWT with ML-DSA for testing.
func GenerateMLDSATestToken(username, machine string, priv *mldsa.PrivateKey) (string, error) {
	header := fmt.Sprintf(`{"alg":"%s","typ":"JWT"}`, mldsa.MLDSA65().String())
	now := time.Now()
	claims, err := json.Marshal(map[string]interface{}{
		"iss":      "github.com/therealpaulgg/ssh-sync",
		"iat":      now.Add(-1 * time.Minute).Unix(),
		"exp":      now.Add(2 * time.Minute).Unix(),
		"username": username,
		"machine":  machine,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal JWT claims: %w", err)
	}

	h := base64.RawURLEncoding.EncodeToString([]byte(header))
	c := base64.RawURLEncoding.EncodeToString(claims)
	signingInput := h + "." + c

	sig, err := priv.Sign(nil, []byte(signingInput), nil)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}
	s := base64.RawURLEncoding.EncodeToString(sig)
	return signingInput + "." + s, nil
}

// GenerateExpiredMLDSATestToken creates an expired ML-DSA JWT for testing.
func GenerateExpiredMLDSATestToken(username, machine string, priv *mldsa.PrivateKey) (string, error) {
	header := fmt.Sprintf(`{"alg":"%s","typ":"JWT"}`, mldsa.MLDSA65().String())
	past := time.Now().Add(-10 * time.Minute)
	claims, err := json.Marshal(map[string]interface{}{
		"iss":      "github.com/therealpaulgg/ssh-sync",
		"iat":      past.Unix(),
		"exp":      past.Add(5 * time.Minute).Unix(),
		"username": username,
		"machine":  machine,
	})
	if err != nil {
		return "", err
	}

	h := base64.RawURLEncoding.EncodeToString([]byte(header))
	c := base64.RawURLEncoding.EncodeToString(claims)
	signingInput := h + "." + c

	sig, err := priv.Sign(nil, []byte(signingInput), nil)
	if err != nil {
		return "", err
	}
	s := base64.RawURLEncoding.EncodeToString(sig)
	return signingInput + "." + s, nil
}
