package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
