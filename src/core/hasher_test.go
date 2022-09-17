package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateHash(t *testing.T) {
	alg := NewDefaultHashAlg()
	path, expectedHash := makeSingleDummyFile(t, &alg.Alg)

	changed, hash, err := UpdateHash(path, alg, false)

	assert.NoError(t, err)
	assert.Equal(t, expectedHash, hash)
	assert.True(t, changed)
	// f, err := os.Open(path)
	// assert.NoError(t, err)

	// // check if hash value is saved to xattr
	// attrHash := GetXattr(f, alg.AttrName)
	// assert.Equal(t, expectedHashValue, attrHash)
}

func TestCalcFileHash(t *testing.T) {
	alg := NewDefaultHashAlg()
	path, expectedHashValue := makeSingleDummyFile(t, &alg.Alg)

	hash, err := CalcFileHash(path, alg)
	assert.NoError(t, err)
	assert.Equal(t, expectedHashValue, hash)
}

func TestCalcFileHash_failed(t *testing.T) {
	alg := NewDefaultHashAlg()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "dummy.txt")

	_, err := CalcFileHash(path, alg)
	assert.Error(t, err)
}

func TestCalcHashString(t *testing.T) {
	alg := NewDefaultHashAlg()
	path, expectedHashValue := makeSingleDummyFile(t, &alg.Alg)

	f, err := os.Open(path)
	assert.NoError(t, err)
	//nolint:errcheck
	defer f.Close()

	hash, err := calcHashString(f, alg)
	assert.NoError(t, err)
	assert.Equal(t, expectedHashValue, hash)
}
