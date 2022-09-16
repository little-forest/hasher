package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcFileHash(t *testing.T) {
	basename := "testdata01"
	hashAlg := NewDefaultHashAlg()

	hash, err := CalcFileHash(fmt.Sprintf("testdata/%s.txt", basename), hashAlg)

	expectedHash := loadHashData(t, basename, hashAlg.AlgName)
	assert.NoError(t, err)
	assert.Equal(t, expectedHash, hash)
}

func TestCalcFileHash_failed(t *testing.T) {
	hashAlg := NewDefaultHashAlg()

	_, err := CalcFileHash("dummy", hashAlg)
	assert.Error(t, err)
}

func TestCalcHashString(t *testing.T) {
	basename := "testdata01"
	hashAlg := NewDefaultHashAlg()

	f := openTestData(t, basename)
	//nolint:errcheck
	defer f.Close()

	hash, err := calcHashString(f, hashAlg)

	expectedHash := loadHashData(t, basename, hashAlg.AlgName)
	assert.NoError(t, err)
	assert.Equal(t, expectedHash, hash)
}
