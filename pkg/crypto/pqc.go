package crypto

import (
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cloudflare/circl/sign/mldsa/mldsa65"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

// KeyType represents the cryptographic algorithm family of a public key.
type KeyType int

const (
	KeyTypeUnknown KeyType = iota
	KeyTypeECDSA
	KeyTypeMLDSA65
)

// DetectKeyType inspects the PEM block type to determine the key algorithm.
func DetectKeyType(pemBytes []byte) KeyType {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return KeyTypeUnknown
	}
	switch block.Type {
	case "PUBLIC KEY":
		return KeyTypeECDSA
	case "MLDSA65 PUBLIC KEY":
		return KeyTypeMLDSA65
	default:
		return KeyTypeUnknown
	}
}

// ParseMLDSA65PublicKey extracts an ML-DSA-65 public key from PEM-encoded bytes.
func ParseMLDSA65PublicKey(pemBytes []byte) (*mldsa65.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	if block.Type != "MLDSA65 PUBLIC KEY" {
		return nil, fmt.Errorf("unexpected PEM block type: %s", block.Type)
	}
	pk := new(mldsa65.PublicKey)
	if err := pk.UnmarshalBinary(block.Bytes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ML-DSA-65 public key: %w", err)
	}
	return pk, nil
}

// ValidatePublicKey validates that the given PEM bytes contain a supported public key.
// Returns the detected KeyType on success.
func ValidatePublicKey(pemBytes []byte) (KeyType, error) {
	kt := DetectKeyType(pemBytes)
	switch kt {
	case KeyTypeECDSA:
		key, err := jwk.ParseKey(pemBytes, jwk.WithPEM(true))
		if err != nil {
			return KeyTypeUnknown, fmt.Errorf("invalid ECDSA key: %w", err)
		}
		if key.KeyType() != jwa.EC {
			return KeyTypeUnknown, errors.New("key is not EC type")
		}
		return KeyTypeECDSA, nil
	case KeyTypeMLDSA65:
		if _, err := ParseMLDSA65PublicKey(pemBytes); err != nil {
			return KeyTypeUnknown, err
		}
		return KeyTypeMLDSA65, nil
	default:
		return KeyTypeUnknown, errors.New("unsupported key type")
	}
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// DetectJWTAlgorithm reads the JWT header to determine the signing algorithm.
func DetectJWTAlgorithm(tokenString string) (string, error) {
	parts := strings.SplitN(tokenString, ".", 3)
	if len(parts) != 3 {
		return "", errors.New("invalid JWT format")
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("failed to decode JWT header: %w", err)
	}
	var header jwtHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return "", fmt.Errorf("failed to parse JWT header: %w", err)
	}
	return header.Alg, nil
}

type jwtClaims struct {
	Username string  `json:"username"`
	Machine  string  `json:"machine"`
	Exp      float64 `json:"exp"`
}

// ExtractJWTClaims manually extracts username and machine claims from a JWT
// without verification. This is used as a fallback when lestrrat-go/jwx
// cannot parse the token (e.g., unrecognized algorithm like MLDSA65).
func ExtractJWTClaims(tokenString string) (username, machine string, err error) {
	parts := strings.SplitN(tokenString, ".", 3)
	if len(parts) != 3 {
		return "", "", errors.New("invalid JWT format")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", fmt.Errorf("failed to decode JWT payload: %w", err)
	}
	var claims jwtClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return "", "", fmt.Errorf("failed to parse JWT claims: %w", err)
	}
	return claims.Username, claims.Machine, nil
}

// VerifyMLDSA65JWT verifies a JWT signed with ML-DSA-65 and checks expiration.
func VerifyMLDSA65JWT(tokenString string, pubKey *mldsa65.PublicKey) error {
	parts := strings.SplitN(tokenString, ".", 3)
	if len(parts) != 3 {
		return errors.New("invalid JWT format")
	}

	// The signed content is the raw "header.payload" string (not decoded)
	signedContent := []byte(parts[0] + "." + parts[1])

	// Decode signature
	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// Verify signature using CIRCL
	if !mldsa65.Verify(pubKey, signedContent, nil, sigBytes) {
		return errors.New("ML-DSA-65 signature verification failed")
	}

	// Decode and validate claims
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("failed to decode payload: %w", err)
	}
	var claims jwtClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return fmt.Errorf("failed to parse claims: %w", err)
	}

	// Check expiration
	if int64(claims.Exp) < time.Now().Unix() {
		return errors.New("token expired")
	}

	return nil
}
