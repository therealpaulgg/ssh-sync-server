package crypto

import (
	"encoding/pem"
	"errors"
	"fmt"

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
