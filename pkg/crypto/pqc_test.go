package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"testing"
	"time"

	"github.com/cloudflare/circl/sign/mldsa/mldsa65"
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

func generateMLDSA65PEM(t *testing.T) ([]byte, *mldsa65.PublicKey, *mldsa65.PrivateKey) {
	t.Helper()
	pub, priv, err := mldsa65.GenerateKey(rand.Reader)
	require.NoError(t, err)
	pubBytes, err := pub.MarshalBinary()
	require.NoError(t, err)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "MLDSA65 PUBLIC KEY", Bytes: pubBytes})
	return pemBytes, pub, priv
}

func signMLDSA65JWT(t *testing.T, priv *mldsa65.PrivateKey, username, machine string, exp time.Time) string {
	t.Helper()
	header := `{"alg":"MLDSA65","typ":"JWT"}`
	claims := fmt.Sprintf(
		`{"iss":"test","iat":%d,"exp":%d,"username":"%s","machine":"%s"}`,
		time.Now().Add(-1*time.Minute).Unix(), exp.Unix(), username, machine,
	)
	h := base64.RawURLEncoding.EncodeToString([]byte(header))
	c := base64.RawURLEncoding.EncodeToString([]byte(claims))
	signingInput := h + "." + c
	sig, err := priv.Sign(nil, []byte(signingInput), nil)
	require.NoError(t, err)
	s := base64.RawURLEncoding.EncodeToString(sig)
	return signingInput + "." + s
}

// --- DetectKeyType ---

func TestDetectKeyType_ECDSA(t *testing.T) {
	pemBytes := generateECDSAPEM(t)
	assert.Equal(t, KeyTypeECDSA, DetectKeyType(pemBytes))
}

func TestDetectKeyType_MLDSA65(t *testing.T) {
	pemBytes, _, _ := generateMLDSA65PEM(t)
	assert.Equal(t, KeyTypeMLDSA65, DetectKeyType(pemBytes))
}

func TestDetectKeyType_Invalid(t *testing.T) {
	assert.Equal(t, KeyTypeUnknown, DetectKeyType([]byte("not a pem")))
}

func TestDetectKeyType_UnknownBlockType(t *testing.T) {
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "SOMETHING ELSE", Bytes: []byte{1, 2, 3}})
	assert.Equal(t, KeyTypeUnknown, DetectKeyType(pemBytes))
}

// --- ParseMLDSA65PublicKey ---

func TestParseMLDSA65PublicKey_Valid(t *testing.T) {
	pemBytes, _, _ := generateMLDSA65PEM(t)
	pk, err := ParseMLDSA65PublicKey(pemBytes)
	require.NoError(t, err)
	assert.NotNil(t, pk)
}

func TestParseMLDSA65PublicKey_WrongPEMType(t *testing.T) {
	pemBytes := generateECDSAPEM(t)
	_, err := ParseMLDSA65PublicKey(pemBytes)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected PEM block type")
}

func TestParseMLDSA65PublicKey_InvalidData(t *testing.T) {
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "MLDSA65 PUBLIC KEY", Bytes: []byte{1, 2, 3}})
	_, err := ParseMLDSA65PublicKey(pemBytes)
	assert.Error(t, err)
}

// --- ValidatePublicKey ---

func TestValidatePublicKey_ECDSA(t *testing.T) {
	pemBytes := generateECDSAPEM(t)
	kt, err := ValidatePublicKey(pemBytes)
	require.NoError(t, err)
	assert.Equal(t, KeyTypeECDSA, kt)
}

func TestValidatePublicKey_MLDSA65(t *testing.T) {
	pemBytes, _, _ := generateMLDSA65PEM(t)
	kt, err := ValidatePublicKey(pemBytes)
	require.NoError(t, err)
	assert.Equal(t, KeyTypeMLDSA65, kt)
}

func TestValidatePublicKey_Invalid(t *testing.T) {
	_, err := ValidatePublicKey([]byte("garbage"))
	assert.Error(t, err)
}

// --- DetectJWTAlgorithm ---

func TestDetectJWTAlgorithm_ES512(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"ES512","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{}`))
	sig := base64.RawURLEncoding.EncodeToString([]byte("fakesig"))
	token := header + "." + payload + "." + sig

	alg, err := DetectJWTAlgorithm(token)
	require.NoError(t, err)
	assert.Equal(t, "ES512", alg)
}

func TestDetectJWTAlgorithm_MLDSA65(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"MLDSA65","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{}`))
	sig := base64.RawURLEncoding.EncodeToString([]byte("fakesig"))
	token := header + "." + payload + "." + sig

	alg, err := DetectJWTAlgorithm(token)
	require.NoError(t, err)
	assert.Equal(t, "MLDSA65", alg)
}

func TestDetectJWTAlgorithm_InvalidFormat(t *testing.T) {
	_, err := DetectJWTAlgorithm("not.a.valid-base64!!!")
	assert.Error(t, err)
}

// --- ExtractJWTClaims ---

func TestExtractJWTClaims(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"MLDSA65"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"username":"alice","machine":"laptop"}`))
	sig := base64.RawURLEncoding.EncodeToString([]byte("sig"))
	token := header + "." + payload + "." + sig

	username, machine, err := ExtractJWTClaims(token)
	require.NoError(t, err)
	assert.Equal(t, "alice", username)
	assert.Equal(t, "laptop", machine)
}

func TestExtractJWTClaims_InvalidFormat(t *testing.T) {
	_, _, err := ExtractJWTClaims("not-a-jwt")
	assert.Error(t, err)
}

// --- VerifyMLDSA65JWT ---

func TestVerifyMLDSA65JWT_Valid(t *testing.T) {
	_, pub, priv := generateMLDSA65PEM(t)
	token := signMLDSA65JWT(t, priv, "user1", "machine1", time.Now().Add(5*time.Minute))
	err := VerifyMLDSA65JWT(token, pub)
	assert.NoError(t, err)
}

func TestVerifyMLDSA65JWT_Expired(t *testing.T) {
	_, pub, priv := generateMLDSA65PEM(t)
	token := signMLDSA65JWT(t, priv, "user1", "machine1", time.Now().Add(-5*time.Minute))
	err := VerifyMLDSA65JWT(token, pub)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestVerifyMLDSA65JWT_BadSignature(t *testing.T) {
	_, _, priv := generateMLDSA65PEM(t)
	token := signMLDSA65JWT(t, priv, "user1", "machine1", time.Now().Add(5*time.Minute))

	// Use a different key to verify
	pub2, _, _ := mldsa65.GenerateKey(rand.Reader)
	err := VerifyMLDSA65JWT(token, pub2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verification failed")
}

func TestVerifyMLDSA65JWT_TamperedPayload(t *testing.T) {
	_, pub, priv := generateMLDSA65PEM(t)
	token := signMLDSA65JWT(t, priv, "user1", "machine1", time.Now().Add(5*time.Minute))

	// Tamper with payload: replace the payload segment
	parts := splitToken(token)
	parts[1] = base64.RawURLEncoding.EncodeToString([]byte(`{"username":"evil","machine":"bad","exp":9999999999}`))
	tampered := parts[0] + "." + parts[1] + "." + parts[2]

	err := VerifyMLDSA65JWT(tampered, pub)
	assert.Error(t, err)
}

func splitToken(token string) [3]string {
	var parts [3]string
	i := 0
	for idx, ch := range token {
		if ch == '.' {
			i++
			continue
		}
		parts[i] += string(token[idx : idx+1])
	}
	return parts
}
