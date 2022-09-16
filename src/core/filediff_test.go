package core

import (
	"encoding/hex"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFileDiff(t *testing.T) {
	basename := "testdata01"
	path := "testdata/testdata01.txt"
	alg := NewDefaultHashAlg()

	d, err := NewFileDiff(path, alg)

	assert.NoError(t, err)
	assert.Equal(t, filepath.Base(path), d.Basename)
	assert.Equal(t, "", d.PairFileName)
	assert.Equal(t, UNKNOWN, d.Status)
	assert.Equal(t, loadHashDataToByteArray(t, basename, alg.AlgName), d.HashValue)
}

func TestNewDirDiff(t *testing.T) {
	alg := NewDefaultHashAlg()

	d, err := NewDirDiff("testdata", alg)
	assert.NoError(t, err)
	assert.NotNil(t, d.Get("testdata01.txt"))
	assert.NotNil(t, d.Get("testdata01.sha1"))
	assert.NotNil(t, d.Get("testdata02.txt"))
	assert.NotNil(t, d.Get("testdata02.sha1"))
	assert.NotNil(t, d.Get("testdata03.txt"))
	assert.NotNil(t, d.Get("testdata03.sha1"))
	assert.Equal(t, 6, d.Count())
}

func loadHashDataToByteArray(t *testing.T, basename string, algName string) []byte {
	hash := loadHashData(t, basename, algName)
	bytes, _ := hex.DecodeString(hash)
	return bytes
}
