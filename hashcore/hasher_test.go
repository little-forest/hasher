package hashcore

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcHash(t *testing.T) {
	alg := NewDefaultHashAlg()
	path, expectedHashValue := makeSingleDummyFile(t, &alg.Alg)

	hash, err := CalcHash(path, alg)
	assert.NoError(t, err)
	assert.Equal(t, expectedHashValue, hash.String())
}

func TestCalcHash_failed(t *testing.T) {
	alg := NewDefaultHashAlg()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "dummy.txt")

	_, err := CalcHash(path, alg)
	assert.Error(t, err)
}
