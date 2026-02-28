package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"
	"testing"
	"time"

	"filippo.io/mldsa"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateECDSAKeyPair(t *testing.T, curve elliptic.Curve) (*ecdsa.PrivateKey, []byte) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	require.NoError(t, err)
	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	require.NoError(t, err)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	return priv, pubPEM
}

func generateECDSAPEM(t *testing.T) []byte {
	t.Helper()
	_, pubPEM := generateECDSAKeyPair(t, elliptic.P256())
	return pubPEM
}

func signECDSAJWT(t *testing.T, priv *ecdsa.PrivateKey, alg jwa.SignatureAlgorithm, exp time.Time) string {
	t.Helper()
	tok, err := jwt.NewBuilder().Expiration(exp).IssuedAt(time.Now()).Build()
	require.NoError(t, err)
	key, err := jwk.FromRaw(priv)
	require.NoError(t, err)
	signed, err := jwt.Sign(tok, jwt.WithKey(alg, key))
	require.NoError(t, err)
	return string(signed)
}

func generateMLDSAPEM(t *testing.T) ([]byte, *mldsa.PublicKey, *mldsa.PrivateKey) {
	t.Helper()
	return generateMLDSAPEMWithParams(t, mldsa.MLDSA65())
}

func generateMLDSAPEMWithParams(t *testing.T, params *mldsa.Parameters) ([]byte, *mldsa.PublicKey, *mldsa.PrivateKey) {
	t.Helper()
	priv, err := mldsa.GenerateKey(params)
	require.NoError(t, err)
	pub := priv.PublicKey()
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "MLDSA PUBLIC KEY", Bytes: pub.Bytes()})
	return pemBytes, pub, priv
}

func signMLDSAJWT(t *testing.T, priv *mldsa.PrivateKey, params *mldsa.Parameters, username, machine string, exp time.Time) string {
	t.Helper()
	header := fmt.Sprintf(`{"alg":"%s","typ":"JWT"}`, params.String())
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

func TestDetectKeyType_ECDSA(t *testing.T) {
	pemBytes := generateECDSAPEM(t)
	assert.Equal(t, KeyTypeECDSA, DetectKeyType(pemBytes))
}

func TestDetectKeyType_MLDSA(t *testing.T) {
	pemBytes, _, _ := generateMLDSAPEM(t)
	assert.Equal(t, KeyTypeMLDSA, DetectKeyType(pemBytes))
}

func TestDetectKeyType_Invalid(t *testing.T) {
	assert.Equal(t, KeyTypeUnknown, DetectKeyType([]byte("not a pem")))
}

func TestDetectKeyType_UnknownBlockType(t *testing.T) {
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "SOMETHING ELSE", Bytes: []byte{1, 2, 3}})
	assert.Equal(t, KeyTypeUnknown, DetectKeyType(pemBytes))
}

func TestParseMLDSAPublicKey_Valid(t *testing.T) {
	pemBytes, _, _ := generateMLDSAPEM(t)
	pk, err := ParseMLDSAPublicKey(pemBytes, mldsa.MLDSA65())
	require.NoError(t, err)
	assert.NotNil(t, pk)
}

func TestParseMLDSAPublicKey_WrongPEMType(t *testing.T) {
	pemBytes := generateECDSAPEM(t)
	_, err := ParseMLDSAPublicKey(pemBytes, mldsa.MLDSA65())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected PEM block type")
}

func TestParseMLDSAPublicKey_InvalidData(t *testing.T) {
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "MLDSA PUBLIC KEY", Bytes: []byte{1, 2, 3}})
	_, err := ParseMLDSAPublicKey(pemBytes, mldsa.MLDSA65())
	assert.Error(t, err)
}

func TestValidatePublicKey_ECDSA(t *testing.T) {
	pemBytes := generateECDSAPEM(t)
	kt, err := ValidatePublicKey(pemBytes)
	require.NoError(t, err)
	assert.Equal(t, KeyTypeECDSA, kt)
}

func TestValidatePublicKey_MLDSA(t *testing.T) {
	pemBytes, _, _ := generateMLDSAPEM(t)
	kt, err := ValidatePublicKey(pemBytes)
	require.NoError(t, err)
	assert.Equal(t, KeyTypeMLDSA, kt)
}

func TestValidatePublicKey_MLDSA44(t *testing.T) {
	pemBytes, _, _ := generateMLDSAPEMWithParams(t, mldsa.MLDSA44())
	kt, err := ValidatePublicKey(pemBytes)
	require.NoError(t, err)
	assert.Equal(t, KeyTypeMLDSA, kt)
}

func TestValidatePublicKey_MLDSA87(t *testing.T) {
	pemBytes, _, _ := generateMLDSAPEMWithParams(t, mldsa.MLDSA87())
	kt, err := ValidatePublicKey(pemBytes)
	require.NoError(t, err)
	assert.Equal(t, KeyTypeMLDSA, kt)
}

func TestValidatePublicKey_MLDSA_UnrecognizedSize(t *testing.T) {
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "MLDSA PUBLIC KEY", Bytes: make([]byte, 42)})
	_, err := ValidatePublicKey(pemBytes)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unrecognized key size")
}

func TestValidatePublicKey_Invalid(t *testing.T) {
	_, err := ValidatePublicKey([]byte("garbage"))
	assert.Error(t, err)
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

func TestDetectJWTAlgorithm_MLDSA65(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"alg":"%s","typ":"JWT"}`, mldsa.MLDSA65().String())))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{}`))
	sig := base64.RawURLEncoding.EncodeToString([]byte("fakesig"))
	token := header + "." + payload + "." + sig

	alg, err := DetectJWTAlgorithm(token)
	require.NoError(t, err)
	assert.Equal(t, mldsa.MLDSA65().String(), alg)
}

func TestDetectJWTAlgorithm_MLDSA44(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"alg":"%s","typ":"JWT"}`, mldsa.MLDSA44().String())))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{}`))
	sig := base64.RawURLEncoding.EncodeToString([]byte("fakesig"))
	token := header + "." + payload + "." + sig

	alg, err := DetectJWTAlgorithm(token)
	require.NoError(t, err)
	assert.Equal(t, mldsa.MLDSA44().String(), alg)
}

func TestDetectJWTAlgorithm_MLDSA87(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"alg":"%s","typ":"JWT"}`, mldsa.MLDSA87().String())))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{}`))
	sig := base64.RawURLEncoding.EncodeToString([]byte("fakesig"))
	token := header + "." + payload + "." + sig

	alg, err := DetectJWTAlgorithm(token)
	require.NoError(t, err)
	assert.Equal(t, mldsa.MLDSA87().String(), alg)
}

func TestDetectJWTAlgorithm_InvalidFormat(t *testing.T) {
	_, err := DetectJWTAlgorithm("not.a.valid-base64!!!")
	assert.Error(t, err)
}

func TestVerifyMLDSAJWT_Valid(t *testing.T) {
	_, pub, priv := generateMLDSAPEM(t)
	token := signMLDSAJWT(t, priv, mldsa.MLDSA65(), "user1", "machine1", time.Now().Add(5*time.Minute))
	err := VerifyMLDSAJWT(token, pub)
	assert.NoError(t, err)
}

func TestVerifyMLDSAJWT_Expired(t *testing.T) {
	_, pub, priv := generateMLDSAPEM(t)
	token := signMLDSAJWT(t, priv, mldsa.MLDSA65(), "user1", "machine1", time.Now().Add(-5*time.Minute))
	err := VerifyMLDSAJWT(token, pub)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestVerifyMLDSAJWT_BadSignature(t *testing.T) {
	_, _, priv := generateMLDSAPEM(t)
	token := signMLDSAJWT(t, priv, mldsa.MLDSA65(), "user1", "machine1", time.Now().Add(5*time.Minute))

	priv2, _ := mldsa.GenerateKey(mldsa.MLDSA65())
	err := VerifyMLDSAJWT(token, priv2.PublicKey())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verification failed")
}

func TestVerifyMLDSAJWT_TamperedPayload(t *testing.T) {
	_, pub, priv := generateMLDSAPEM(t)
	token := signMLDSAJWT(t, priv, mldsa.MLDSA65(), "user1", "machine1", time.Now().Add(5*time.Minute))

	parts := strings.SplitN(token, ".", 3)
	require.Len(t, parts, 3)
	parts[1] = base64.RawURLEncoding.EncodeToString([]byte(`{"username":"evil","machine":"bad","exp":9999999999}`))
	tampered := parts[0] + "." + parts[1] + "." + parts[2]

	err := VerifyMLDSAJWT(tampered, pub)
	assert.Error(t, err)
}

func TestVerifyJWT_ES256_Valid(t *testing.T) {
	priv, pubPEM := generateECDSAKeyPair(t, elliptic.P256())
	token := signECDSAJWT(t, priv, jwa.ES256, time.Now().Add(5*time.Minute))
	err := VerifyJWT(token, jwa.ES256.String(), pubPEM)
	assert.NoError(t, err)
}

func TestVerifyJWT_ES512_Valid(t *testing.T) {
	priv, pubPEM := generateECDSAKeyPair(t, elliptic.P521())
	token := signECDSAJWT(t, priv, jwa.ES512, time.Now().Add(5*time.Minute))
	err := VerifyJWT(token, jwa.ES512.String(), pubPEM)
	assert.NoError(t, err)
}

func TestVerifyJWT_ECDSA_Expired(t *testing.T) {
	priv, pubPEM := generateECDSAKeyPair(t, elliptic.P256())
	token := signECDSAJWT(t, priv, jwa.ES256, time.Now().Add(-5*time.Minute))
	err := VerifyJWT(token, jwa.ES256.String(), pubPEM)
	assert.Error(t, err)
}

func TestVerifyJWT_ECDSA_WrongKey(t *testing.T) {
	priv, _ := generateECDSAKeyPair(t, elliptic.P256())
	_, otherPubPEM := generateECDSAKeyPair(t, elliptic.P256())
	token := signECDSAJWT(t, priv, jwa.ES256, time.Now().Add(5*time.Minute))
	err := VerifyJWT(token, jwa.ES256.String(), otherPubPEM)
	assert.Error(t, err)
}

func TestVerifyJWT_MLDSA65_Valid(t *testing.T) {
	pemBytes, _, priv := generateMLDSAPEM(t)
	token := signMLDSAJWT(t, priv, mldsa.MLDSA65(), "user1", "machine1", time.Now().Add(5*time.Minute))
	err := VerifyJWT(token, mldsa.MLDSA65().String(), pemBytes)
	assert.NoError(t, err)
}

func TestVerifyJWT_MLDSA_WrongAlg(t *testing.T) {
	pemBytes, _, priv := generateMLDSAPEM(t) // MLDSA65 key
	token := signMLDSAJWT(t, priv, mldsa.MLDSA65(), "user1", "machine1", time.Now().Add(5*time.Minute))
	// Claim it's MLDSA44 — key size mismatch will cause parsing to fail
	err := VerifyJWT(token, mldsa.MLDSA44().String(), pemBytes)
	assert.Error(t, err)
}

func TestVerifyJWT_UnsupportedAlg(t *testing.T) {
	_, pubPEM := generateECDSAKeyPair(t, elliptic.P256())
	err := VerifyJWT("a.b.c", "RS256", pubPEM)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported JWT algorithm")
}

func TestVerifyJWT_MLDSA87_Valid(t *testing.T) {
	pemBytes, _, priv := generateMLDSAPEMWithParams(t, mldsa.MLDSA87())
	token := signMLDSAJWT(t, priv, mldsa.MLDSA87(), "user1", "machine1", time.Now().Add(5*time.Minute))
	err := VerifyJWT(token, mldsa.MLDSA87().String(), pemBytes)
	assert.NoError(t, err)
}

func TestVerifyJWT_MLDSA_FakeAlg(t *testing.T) {
	pemBytes, _, _ := generateMLDSAPEM(t)
	err := VerifyJWT("a.b.c", "ML-DSA-69", pemBytes)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported JWT algorithm")
}

func TestVerifyJWT_ECDSA_InvalidKey(t *testing.T) {
	err := VerifyJWT("a.b.c", jwa.ES256.String(), []byte("not a pem"))
	assert.Error(t, err)
}

func TestDetectJWTAlgorithm_TooFewParts(t *testing.T) {
	_, err := DetectJWTAlgorithm("onlytwoparts.here")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JWT format")
}

func TestDetectJWTAlgorithm_InvalidHeaderBase64(t *testing.T) {
	_, err := DetectJWTAlgorithm("!!!invalid-base64!!!.payload.sig")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode JWT header")
}

