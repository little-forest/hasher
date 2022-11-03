package core

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateHash(t *testing.T) {
	alg := NewDefaultHashAlg()
	path, expectedHash := makeSingleDummyFile(t, &alg.Alg)

	changed, hash, err := UpdateHash2(path, alg, false)

	assert.NoError(t, err)
	assert.Equal(t, expectedHash, hash.String())
	assert.True(t, changed)
	// f, err := os.Open(path)
	// assert.NoError(t, err)

	// // check if hash value is saved to xattr
	// attrHash := GetXattr(f, alg.AttrName)
	// assert.Equal(t, expectedHashValue, attrHash)
}

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
