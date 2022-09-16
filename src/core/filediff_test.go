package core

import (
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
