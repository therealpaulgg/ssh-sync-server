package crypto

import (
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	"filippo.io/mldsa"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

type KeyType int

const (
	KeyTypeUnknown KeyType = iota
	KeyTypeECDSA
	KeyTypeMLDSA
)

func DetectKeyType(pemBytes []byte) KeyType {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return KeyTypeUnknown
	}
	switch block.Type {
	case "PUBLIC KEY":
		return KeyTypeECDSA
	case "MLDSA PUBLIC KEY":
		return KeyTypeMLDSA
	default:
		return KeyTypeUnknown
	}
}

func ParseMLDSAPublicKey(pemBytes []byte) (*mldsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	if block.Type != "MLDSA PUBLIC KEY" {
		return nil, fmt.Errorf("unexpected PEM block type: %s", block.Type)
	}
	pk, err := mldsa.NewPublicKey(mldsa.MLDSA65(), block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ML-DSA public key: %w", err)
	}
	return pk, nil
}

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
	case KeyTypeMLDSA:
		if _, err := ParseMLDSAPublicKey(pemBytes); err != nil {
			return KeyTypeUnknown, err
		}
		return KeyTypeMLDSA, nil
	default:
		return KeyTypeUnknown, errors.New("unsupported key type")
	}
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

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

func VerifyMLDSAJWT(tokenString string, pubKey *mldsa.PublicKey) error {
	parts := strings.SplitN(tokenString, ".", 3)
	if len(parts) != 3 {
		return errors.New("invalid JWT format")
	}

	signedContent := []byte(parts[0] + "." + parts[1])

	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	if err := mldsa.Verify(pubKey, signedContent, sigBytes, nil); err != nil {
		return errors.New("ML-DSA signature verification failed")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("failed to decode payload: %w", err)
	}
	var claims jwtClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return fmt.Errorf("failed to parse claims: %w", err)
	}

	if int64(claims.Exp) <= time.Now().Unix() {
		return errors.New("token expired")
	}

	return nil
}
