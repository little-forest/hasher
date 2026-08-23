package hasher

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/little-forest/hasher/hashcore"
	"github.com/stretchr/testify/assert"
)

func TestNewFileDiff(t *testing.T) {
	alg := hashcore.NewDefaultHashAlg()
	path, expectedHashValue := makeSingleDummyFile(t, &alg.Alg)
	expectedHashBytes, _ := hex.DecodeString(expectedHashValue)

	d, err := NewFileDiff(path, alg)
	assert.NoError(t, err)

	info, err := os.Stat(path)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, filepath.Base(path), d.Basename)
	assert.Equal(t, "", d.PairFileName)
	assert.Equal(t, UNKNOWN, d.Status)
	assert.True(t, info.ModTime().Equal(d.ModTime))
	assert.Equal(t, expectedHashBytes, d.HashValue)
	assert.Nil(t, d.Parent)
}

func TestFileDiffCompare(t *testing.T) {
	// TODO:
}
