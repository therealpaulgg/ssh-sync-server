package crypto

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"filippo.io/mldsa"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

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

// VerifyJWT verifies the signature of a JWT using the stored public key PEM.
func VerifyJWT(tokenString, alg string, publicKeyPEM []byte) error {
	switch alg {
	case "ES256", "ES512":
		key, err := jwk.ParseKey(publicKeyPEM, jwk.WithPEM(true))
		if err != nil {
			return fmt.Errorf("parsing EC public key: %w", err)
		}
		if _, err := jwt.ParseString(tokenString, jwt.WithKey(jwa.SignatureAlgorithm(alg), key)); err != nil {
			return fmt.Errorf("EC JWT verification failed: %w", err)
		}
	case mldsa.MLDSA44().String(), mldsa.MLDSA65().String(), mldsa.MLDSA87().String():
		mldsaAlg, err := MLDSAAlgorithmFromString(alg)
		if err != nil {
			return err
		}
		pubKey, err := ParseMLDSAPublicKey(publicKeyPEM, mldsaAlg)
		if err != nil {
			return fmt.Errorf("parsing ML-DSA public key: %w", err)
		}
		if err := VerifyMLDSAJWT(tokenString, pubKey); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported JWT algorithm: %s", alg)
	}
	return nil
}
