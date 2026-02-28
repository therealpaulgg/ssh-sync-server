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
)

type jwtExpClaims struct {
	Exp float64 `json:"exp"`
}

// MLDSAAlgorithmFromString maps a JOSE algorithm identifier to its *mldsa.Parameters.
// See draft-ietf-cose-dilithium: https://datatracker.ietf.org/doc/draft-ietf-cose-dilithium/
func MLDSAAlgorithmFromString(algStr string) (*mldsa.Parameters, error) {
	switch algStr {
	case mldsa.MLDSA44().String():
		return mldsa.MLDSA44(), nil
	case mldsa.MLDSA65().String():
		return mldsa.MLDSA65(), nil
	case mldsa.MLDSA87().String():
		return mldsa.MLDSA87(), nil
	default:
		return nil, fmt.Errorf("unsupported ML-DSA algorithm: %s", algStr)
	}
}

func ParseMLDSAPublicKey(pemBytes []byte, algorithm *mldsa.Parameters) (*mldsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	if block.Type != "MLDSA PUBLIC KEY" {
		return nil, fmt.Errorf("unexpected PEM block type: %s", block.Type)
	}
	pk, err := mldsa.NewPublicKey(algorithm, block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ML-DSA public key: %w", err)
	}
	return pk, nil
}

// VerifyMLDSAJWT verifies a JWT signed with an ML-DSA variant.
// See draft-ietf-cose-dilithium: https://datatracker.ietf.org/doc/draft-ietf-cose-dilithium/
func VerifyMLDSAJWT(tokenString string, pubKey *mldsa.PublicKey) error {
	parts := strings.SplitN(tokenString, ".", 3)
	if len(parts) != 3 {
		return errors.New("invalid JWT format")
	}

	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	if err := mldsa.Verify(pubKey, []byte(parts[0]+"."+parts[1]), sigBytes, nil); err != nil {
		return errors.New("ML-DSA signature verification failed")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("failed to decode payload: %w", err)
	}
	var claims jwtExpClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return fmt.Errorf("failed to parse claims: %w", err)
	}
	if int64(claims.Exp) <= time.Now().Unix() {
		return errors.New("token expired")
	}

	return nil
}
