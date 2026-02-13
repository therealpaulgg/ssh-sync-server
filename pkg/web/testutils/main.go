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

	"github.com/cloudflare/circl/sign/mldsa/mldsa65"
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

// GenerateMLDSA65TestKeys generates an ML-DSA-65 keypair via CIRCL.
func GenerateMLDSA65TestKeys() (*mldsa65.PublicKey, *mldsa65.PrivateKey, error) {
	pub, priv, err := mldsa65.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return pub, priv, nil
}

// EncodeMLDSA65ToPem PEM-encodes an ML-DSA-65 public key.
func EncodeMLDSA65ToPem(pub *mldsa65.PublicKey) ([]byte, error) {
	pubBytes, err := pub.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "MLDSA65 PUBLIC KEY",
		Bytes: pubBytes,
	}), nil
}

// GenerateMLDSA65TestToken creates and signs a JWT with ML-DSA-65 for testing.
func GenerateMLDSA65TestToken(username, machine string, priv *mldsa65.PrivateKey) (string, error) {
	header := `{"alg":"MLDSA65","typ":"JWT"}`
	now := time.Now()
	claims := fmt.Sprintf(
		`{"iss":"github.com/therealpaulgg/ssh-sync","iat":%d,"exp":%d,"username":"%s","machine":"%s"}`,
		now.Add(-1*time.Minute).Unix(),
		now.Add(2*time.Minute).Unix(),
		username,
		machine,
	)

	h := base64.RawURLEncoding.EncodeToString([]byte(header))
	c := base64.RawURLEncoding.EncodeToString([]byte(claims))
	signingInput := h + "." + c

	sig, err := priv.Sign(nil, []byte(signingInput), nil)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}
	s := base64.RawURLEncoding.EncodeToString(sig)
	return signingInput + "." + s, nil
}

// GenerateExpiredMLDSA65TestToken creates an expired ML-DSA-65 JWT for testing.
func GenerateExpiredMLDSA65TestToken(username, machine string, priv *mldsa65.PrivateKey) (string, error) {
	header := `{"alg":"MLDSA65","typ":"JWT"}`
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
