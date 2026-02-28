package crypto

import (
	"testing"

	"filippo.io/mldsa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMLDSAAlgorithmFromString_Valid(t *testing.T) {
	for _, params := range []*mldsa.Parameters{mldsa.MLDSA44(), mldsa.MLDSA65(), mldsa.MLDSA87()} {
		result, err := MLDSAAlgorithmFromString(params.String())
		require.NoError(t, err)
		assert.Equal(t, params, result)
	}
}

func TestMLDSAAlgorithmFromString_Unsupported(t *testing.T) {
	_, err := MLDSAAlgorithmFromString("ML-DSA-69")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported ML-DSA algorithm")
}
