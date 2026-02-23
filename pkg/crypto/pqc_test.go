package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateECDSAPEM(t *testing.T) []byte {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	require.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
}

func generateEncapsulationKeyPEM(t *testing.T) []byte {
	t.Helper()
	keyBytes := make([]byte, 1184) // ML-KEM-768 encapsulation key size
	_, err := rand.Read(keyBytes)
	require.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "MLKEM768 ENCAPSULATION KEY", Bytes: keyBytes})
}

// --- ParsePublicKeyPEM ---

func TestParsePublicKeyPEM_ECDSAOnly(t *testing.T) {
	ecPEM := generateECDSAPEM(t)
	parsed, err := ParsePublicKeyPEM(ecPEM)
	require.NoError(t, err)
	assert.NotNil(t, parsed.SigningKey)
	assert.Nil(t, parsed.EncapsulationKey)
}

func TestParsePublicKeyPEM_Hybrid(t *testing.T) {
	ecPEM := generateECDSAPEM(t)
	encapPEM := generateEncapsulationKeyPEM(t)
	combined := append(ecPEM, encapPEM...)

	parsed, err := ParsePublicKeyPEM(combined)
	require.NoError(t, err)
	assert.NotNil(t, parsed.SigningKey)
	assert.NotNil(t, parsed.EncapsulationKey)
}

func TestParsePublicKeyPEM_NoPublicKey(t *testing.T) {
	encapPEM := generateEncapsulationKeyPEM(t)
	_, err := ParsePublicKeyPEM(encapPEM)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no EC public key found")
}

func TestParsePublicKeyPEM_InvalidPEM(t *testing.T) {
	_, err := ParsePublicKeyPEM([]byte("not a pem"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no EC public key found")
}

func TestParsePublicKeyPEM_UnknownBlockType(t *testing.T) {
	badPEM := pem.EncodeToMemory(&pem.Block{Type: "SOMETHING ELSE", Bytes: []byte{1, 2, 3}})
	_, err := ParsePublicKeyPEM(badPEM)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected PEM block type")
}

func TestParsePublicKeyPEM_WrongEncapsulationKeySize(t *testing.T) {
	ecPEM := generateECDSAPEM(t)
	badEncapPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "MLKEM768 ENCAPSULATION KEY",
		Bytes: []byte{1, 2, 3}, // wrong size
	})
	combined := append(ecPEM, badEncapPEM...)
	_, err := ParsePublicKeyPEM(combined)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wrong size")
}

func TestParsePublicKeyPEM_NotECKey(t *testing.T) {
	badPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: []byte{1, 2, 3}})
	_, err := ParsePublicKeyPEM(badPEM)
	assert.Error(t, err)
}

// --- ValidateEncapsulationKeyPEM ---

func TestValidateEncapsulationKeyPEM_Valid(t *testing.T) {
	encapPEM := generateEncapsulationKeyPEM(t)
	err := ValidateEncapsulationKeyPEM(encapPEM)
	assert.NoError(t, err)
}

func TestValidateEncapsulationKeyPEM_Empty(t *testing.T) {
	err := ValidateEncapsulationKeyPEM(nil)
	assert.NoError(t, err)
	err = ValidateEncapsulationKeyPEM([]byte{})
	assert.NoError(t, err)
}

func TestValidateEncapsulationKeyPEM_WrongBlockType(t *testing.T) {
	badPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: make([]byte, 1184)})
	err := ValidateEncapsulationKeyPEM(badPEM)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected PEM block type")
}

func TestValidateEncapsulationKeyPEM_WrongSize(t *testing.T) {
	badPEM := pem.EncodeToMemory(&pem.Block{Type: "MLKEM768 ENCAPSULATION KEY", Bytes: []byte{1, 2, 3}})
	err := ValidateEncapsulationKeyPEM(badPEM)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wrong size")
}

func TestValidateEncapsulationKeyPEM_InvalidPEM(t *testing.T) {
	err := ValidateEncapsulationKeyPEM([]byte("not a pem"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode")
}

// --- DetectJWTAlgorithm ---

func TestDetectJWTAlgorithm_ES256(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"ES256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{}`))
	sig := base64.RawURLEncoding.EncodeToString([]byte("fakesig"))
	token := header + "." + payload + "." + sig

	alg, err := DetectJWTAlgorithm(token)
	require.NoError(t, err)
	assert.Equal(t, "ES256", alg)
}

func TestDetectJWTAlgorithm_ES512(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"ES512","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{}`))
	sig := base64.RawURLEncoding.EncodeToString([]byte("fakesig"))
	token := header + "." + payload + "." + sig

	alg, err := DetectJWTAlgorithm(token)
	require.NoError(t, err)
	assert.Equal(t, "ES512", alg)
}

func TestDetectJWTAlgorithm_InvalidFormat(t *testing.T) {
	_, err := DetectJWTAlgorithm("not.a.valid-base64!!!")
	assert.Error(t, err)
}

func TestDetectJWTAlgorithm_NotEnoughParts(t *testing.T) {
	_, err := DetectJWTAlgorithm("only-one-part")
	assert.Error(t, err)
}
