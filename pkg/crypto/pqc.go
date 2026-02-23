package crypto

import (
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

const mlkem768EncapsulationKeySize = 1184

// ParsedKeys holds the signing key and optional encapsulation key
// extracted from a (possibly multi-block) PEM upload.
type ParsedKeys struct {
	SigningKey        []byte // PEM-encoded EC public key (always present)
	EncapsulationKey  []byte // PEM-encoded ML-KEM-768 encapsulation key (nil for legacy)
}

// ParsePublicKeyPEM parses a PEM file that may contain one or two blocks:
//   - "PUBLIC KEY" (ECDSA, required) — used for JWT verification
//   - "MLKEM768 ENCAPSULATION KEY" (optional) — stored for hybrid key exchange
//
// Legacy users send only the PUBLIC KEY block. Hybrid users send both.
func ParsePublicKeyPEM(pemData []byte) (*ParsedKeys, error) {
	result := &ParsedKeys{}
	rest := pemData

	for {
		block, remaining := pem.Decode(rest)
		if block == nil {
			break
		}
		switch block.Type {
		case "PUBLIC KEY":
			// Validate it's actually an EC key
			singlePEM := pem.EncodeToMemory(block)
			key, err := jwk.ParseKey(singlePEM, jwk.WithPEM(true))
			if err != nil {
				return nil, fmt.Errorf("invalid EC public key: %w", err)
			}
			if key.KeyType() != jwa.EC {
				return nil, errors.New("public key is not EC type")
			}
			result.SigningKey = singlePEM
		case "MLKEM768 ENCAPSULATION KEY":
			if len(block.Bytes) != mlkem768EncapsulationKeySize {
				return nil, fmt.Errorf("ML-KEM-768 encapsulation key has wrong size: got %d, want %d", len(block.Bytes), mlkem768EncapsulationKeySize)
			}
			result.EncapsulationKey = pem.EncodeToMemory(block)
		default:
			return nil, fmt.Errorf("unexpected PEM block type: %s", block.Type)
		}
		rest = remaining
	}

	if result.SigningKey == nil {
		return nil, errors.New("no EC public key found in PEM data")
	}
	return result, nil
}

// ValidateEncapsulationKeyPEM validates a standalone ML-KEM-768 encapsulation key PEM.
func ValidateEncapsulationKeyPEM(pemBytes []byte) error {
	if len(pemBytes) == 0 {
		return nil // nil/empty is valid (legacy user)
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return errors.New("failed to decode encapsulation key PEM")
	}
	if block.Type != "MLKEM768 ENCAPSULATION KEY" {
		return fmt.Errorf("unexpected PEM block type for encapsulation key: %s", block.Type)
	}
	if len(block.Bytes) != mlkem768EncapsulationKeySize {
		return fmt.Errorf("ML-KEM-768 encapsulation key has wrong size: got %d, want %d", len(block.Bytes), mlkem768EncapsulationKeySize)
	}
	return nil
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
