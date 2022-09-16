package core

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFileDiff(t *testing.T) {
	basename := "testdata01"
	path := "testdata/testdata01.txt"
	alg := NewDefaultHashAlg()

	d, err := NewFileDiff(path, alg)

	assert.NoError(t, err)
	assert.Equal(t, path, d.FileName)
	assert.Equal(t, "", d.PairFileName)
	assert.Equal(t, loadHashDataToByteArray(t, basename, alg.AlgName), d.HashValue)
}

func loadHashDataToByteArray(t *testing.T, basename string, algName string) []byte {
	hash := loadHashData(t, basename, algName)
	bytes, _ := hex.DecodeString(hash)
	return bytes
}
