package core

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDirDiff(t *testing.T) {
	alg := NewDefaultHashAlg()
	basedir := t.TempDir()

	path1 := filepath.Join(basedir, "test01")
	path2 := filepath.Join(basedir, "test02")
	path3 := filepath.Join(basedir, "test03")

	makeDummyFile(t, path1, &alg.Alg)
	makeDummyFile(t, path2, &alg.Alg)
	makeDummyFile(t, path3, &alg.Alg)

	d, err := NewDirDiff(basedir, alg)
	assert.NoError(t, err)
	assert.NotNil(t, d.Get("test01"))
	assert.NotNil(t, d.Get("test02"))
	assert.NotNil(t, d.Get("test03"))
	assert.Equal(t, 3, d.Count())
}

func TestDummy(t *testing.T) {
	alg := NewDefaultHashAlg()

	meDir := t.TempDir()
	otherDir := t.TempDir()

	// me
	makeDummyFile(t, filepath.Join(meDir, "01.txt"), &alg.Alg)
	makeDummyFile(t, filepath.Join(meDir, "02.txt"), &alg.Alg)
	makeDummyFile(t, filepath.Join(meDir, "03.txt"), &alg.Alg)

	// other
	//   same file
	copyFile(t, meDir, otherDir, "01.txt")
	//   changed file
	makeDummyFile(t, filepath.Join(otherDir, "02.txt"), &alg.Alg)
	//   removed file
	makeDummyFile(t, filepath.Join(otherDir, "04.txt"), &alg.Alg)

	me, err := NewDirDiff(meDir, alg)
	assert.NoError(t, err)

	other, err := NewDirDiff(otherDir, alg)
	assert.NoError(t, err)

	assert.Equal(t, 3, me.Count())
	assert.Equal(t, 3, other.Count())
}
