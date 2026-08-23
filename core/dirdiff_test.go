package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDirDiff(t *testing.T) {
	alg := NewDefaultHashAlg()
	basedir, _ := prepareDirDiffTest_01(t, alg)

	d, err := NewDirDiff(basedir, alg)
	assert.NoError(t, err)
	assert.NotNil(t, d.Get("test01"))
	assert.Equal(t, d, d.Get("test01").Parent)
	assert.NotNil(t, d.Get("test02"))
	assert.Equal(t, d, d.Get("test02").Parent)
	assert.NotNil(t, d.Get("test03"))
	assert.Equal(t, d, d.Get("test03").Parent)
	assert.Equal(t, 3, d.Count())
}

func TestGetChildren(t *testing.T) {
	alg := NewDefaultHashAlg()
	basedir, _ := prepareDirDiffTest_01(t, alg)

	d, err := NewDirDiff(basedir, alg)
	assert.NoError(t, err)
	children := d.GetChildren()

	assert.Equal(t, 3, len(children))
}

func TestDirDiffCompare_01(t *testing.T) {
	alg := NewDefaultHashAlg()
	meDir, otherDir := prepareDirDiffTest_01(t, alg)

	me, err := NewDirDiff(meDir, alg)
	assert.NoError(t, err)
	other, err := NewDirDiff(otherDir, alg)
	assert.NoError(t, err)

	me.Compare(other)

	// assert
	files := me.GetSortedChildren()
	assert.Equal(t, 3, len(files))
	assertFileDiff(t, "test01", SAME, "test01", files[0])
	assertFileDiff(t, "test02", SAME, "test02", files[1])
	assertFileDiff(t, "test03", SAME, "test03", files[2])
}

func TestDirDiffCompare_02(t *testing.T) {
	alg := NewDefaultHashAlg()
	meDir, otherDir := prepareDirDiffTest_02(t, alg)

	me, err := NewDirDiff(meDir, alg)
	assert.NoError(t, err)
	other, err := NewDirDiff(otherDir, alg)
	assert.NoError(t, err)

	me.Compare(other)

	// assert
	files := me.GetSortedChildren()
	assert.Equal(t, 6, len(files))
	assertFileDiff(t, "test01", SAME, "test01", files[0])
	assertFileDiff(t, "test02", NOT_SAME_NEW, "test02", files[1])
	assertFileDiff(t, "test03", NOT_SAME_OLD, "test03", files[2])
	assertFileDiff(t, "test04", NOT_SAME, "test04", files[3])
	assertFileDiff(t, "test05", ADDED, "", files[4])
	assertFileDiff(t, "test06", REMOVED, "", files[5])
}

func TestDirDiffCompare_03(t *testing.T) {
	alg := NewDefaultHashAlg()
	meDir, otherDir := prepareDirDiffTest_03(t, alg)

	me, err := NewDirDiff(meDir, alg)
	assert.NoError(t, err)
	other, err := NewDirDiff(otherDir, alg)
	assert.NoError(t, err)

	me.Compare(other)

	// assert
	files := me.GetSortedChildren()
	assert.Equal(t, 3, len(files))
	assertFileDiff(t, "test01", ADDED, "", files[0])
	assertFileDiff(t, "test02", ADDED, "", files[1])
	assertFileDiff(t, "test03", ADDED, "", files[2])
}

func TestDirDiffCompare_04(t *testing.T) {
	alg := NewDefaultHashAlg()
	meDir, otherDir := prepareDirDiffTest_04(t, alg)

	me, err := NewDirDiff(meDir, alg)
	assert.NoError(t, err)
	other, err := NewDirDiff(otherDir, alg)
	assert.NoError(t, err)

	me.Compare(other)

	// assert
	files := me.GetSortedChildren()
	assert.Equal(t, 3, len(files))
	assertFileDiff(t, "test01", REMOVED, "", files[0])
	assertFileDiff(t, "test02", REMOVED, "", files[1])
	assertFileDiff(t, "test03", REMOVED, "", files[2])
}

func TestDirDiffCompare_05(t *testing.T) {
	alg := NewDefaultHashAlg()
	meDir, otherDir := prepareDirDiffTest_05(t, alg)

	me, err := NewDirDiff(meDir, alg)
	assert.NoError(t, err)
	other, err := NewDirDiff(otherDir, alg)
	assert.NoError(t, err)

	me.Compare(other)

	// assert
	files := me.GetSortedChildren()
	assert.Equal(t, 3, len(files))
	assertFileDiff(t, "test0A", RENAMED, "test01", files[0])
	assertFileDiff(t, "test0B", RENAMED, "test02", files[1])
	assertFileDiff(t, "test0C", RENAMED, "test03", files[2])
}

func TestDirDiffCompare_06(t *testing.T) {
	alg := NewDefaultHashAlg()
	meDir, otherDir := prepareDirDiffTest_06(t, alg)

	me, err := NewDirDiff(meDir, alg)
	assert.NoError(t, err)
	other, err := NewDirDiff(otherDir, alg)
	assert.NoError(t, err)

	me.Compare(other)

	// assert
	files := me.GetSortedChildren()
	assert.Equal(t, 6, len(files))
	assertFileDiff(t, "test01", SAME, "test01", files[0])
	assertFileDiff(t, "test02", NOT_SAME_NEW, "test02", files[1])
	assertFileDiff(t, "test04", NOT_SAME, "test04", files[2])
	assertFileDiff(t, "test05", ADDED, "", files[3])
	assertFileDiff(t, "test06", REMOVED, "", files[4])
	assertFileDiff(t, "test07", RENAMED, "test03", files[5])
}

// DirDiff test pattern1
//
//	[=] test01 <-> test01
//	[=] test02 <-> test02
//	[=] test03 <-> test03
func prepareDirDiffTest_01(t *testing.T, alg *HashAlg) (string, string) {
	t.Helper()

	meDir := t.TempDir()
	otherDir := t.TempDir()

	path1 := filepath.Join(meDir, "test01")
	path2 := filepath.Join(meDir, "test02")
	path3 := filepath.Join(meDir, "test03")

	// me
	makeDummyFile(t, path1, &alg.Alg)
	makeDummyFile(t, path2, &alg.Alg)
	makeDummyFile(t, path3, &alg.Alg)

	copyFile(t, meDir, otherDir, "test01")
	copyFile(t, meDir, otherDir, "test02")
	copyFile(t, meDir, otherDir, "test03")

	return meDir, otherDir
}

// DirDiff test pattern2
//
//	[=] test01 <-> test01
//	[>] test02 <-> test02
//	[<] test03 <-> test03
//	[~] test04 <-> test04
//	[+] test05 <->
//	[-]        <-> test06
func prepareDirDiffTest_02(t *testing.T, alg *HashAlg) (string, string) {
	t.Helper()

	meDir := t.TempDir()
	otherDir := t.TempDir()

	mPath1 := filepath.Join(meDir, "test01")
	mPath2 := filepath.Join(meDir, "test02")
	mPath3 := filepath.Join(meDir, "test03")
	mPath4 := filepath.Join(meDir, "test04")
	mPath5 := filepath.Join(meDir, "test05")

	oPath2 := filepath.Join(otherDir, "test02")
	oPath3 := filepath.Join(otherDir, "test03")
	oPath4 := filepath.Join(otherDir, "test04")
	oPath6 := filepath.Join(otherDir, "test06")

	// me
	makeDummyFile(t, mPath1, &alg.Alg)
	makeDummyFile(t, mPath2, &alg.Alg)
	makeDummyFile(t, mPath3, &alg.Alg)
	makeDummyFile(t, mPath4, &alg.Alg)
	makeDummyFile(t, mPath5, &alg.Alg)

	// other
	copyFile(t, meDir, otherDir, "test01")
	makeDummyFile(t, oPath2, &alg.Alg)
	touchDelta(t, mPath2, oPath2, time.Minute)
	makeDummyFile(t, oPath3, &alg.Alg)
	touchDelta(t, mPath3, oPath3, -time.Minute)
	makeDummyFile(t, oPath4, &alg.Alg)
	touchDelta(t, mPath4, oPath4, 0)
	makeDummyFile(t, oPath6, &alg.Alg)

	return meDir, otherDir
}

// DirDiff test pattern3
//
//	[+] test01 <->
//	[+] test02 <->
//	[+] test03 <->
func prepareDirDiffTest_03(t *testing.T, alg *HashAlg) (string, string) {
	t.Helper()

	meDir := t.TempDir()
	otherDir := t.TempDir()

	path1 := filepath.Join(meDir, "test01")
	path2 := filepath.Join(meDir, "test02")
	path3 := filepath.Join(meDir, "test03")

	// me
	makeDummyFile(t, path1, &alg.Alg)
	makeDummyFile(t, path2, &alg.Alg)
	makeDummyFile(t, path3, &alg.Alg)

	return meDir, otherDir
}

// DirDiff test pattern4
//
//	[-]      <-> test01
//	[-]      <-> test02
//	[-]      <-> test03
func prepareDirDiffTest_04(t *testing.T, alg *HashAlg) (string, string) {
	t.Helper()

	meDir := t.TempDir()
	otherDir := t.TempDir()

	path1 := filepath.Join(otherDir, "test01")
	path2 := filepath.Join(otherDir, "test02")
	path3 := filepath.Join(otherDir, "test03")

	// other
	makeDummyFile(t, path1, &alg.Alg)
	makeDummyFile(t, path2, &alg.Alg)
	makeDummyFile(t, path3, &alg.Alg)

	return meDir, otherDir
}

// DirDiff test pattern5
//
//	[R] test0A <-> test01
//	[R] test0B <-> test02
//	[R] test0C <-> test03
func prepareDirDiffTest_05(t *testing.T, alg *HashAlg) (string, string) {
	t.Helper()

	meDir := t.TempDir()
	otherDir := t.TempDir()

	path1 := filepath.Join(otherDir, "test01")
	path2 := filepath.Join(otherDir, "test02")
	path3 := filepath.Join(otherDir, "test03")

	// other
	makeDummyFile(t, path1, &alg.Alg)
	makeDummyFile(t, path2, &alg.Alg)
	makeDummyFile(t, path3, &alg.Alg)

	// me
	copyFile(t, otherDir, meDir, "test01")
	// nolint:errcheck
	os.Rename(filepath.Join(meDir, "test01"), filepath.Join(meDir, "test0A"))
	copyFile(t, otherDir, meDir, "test02")
	// nolint:errcheck
	os.Rename(filepath.Join(meDir, "test02"), filepath.Join(meDir, "test0B"))
	copyFile(t, otherDir, meDir, "test03")
	// nolint:errcheck
	os.Rename(filepath.Join(meDir, "test03"), filepath.Join(meDir, "test0C"))

	return meDir, otherDir
}

// DirDiff test pattern6
//
//	[=] test01 <-> test01
//	[>] test02 <-> test02
//	[R] test07 <-> test03
//	[~] test04 <-> test04
//	[+] test05 <->
//	[-]        <-> test06
func prepareDirDiffTest_06(t *testing.T, alg *HashAlg) (string, string) {
	t.Helper()

	meDir := t.TempDir()
	otherDir := t.TempDir()

	mPath1 := filepath.Join(meDir, "test01")
	mPath2 := filepath.Join(meDir, "test02")
	mPath4 := filepath.Join(meDir, "test04")
	mPath5 := filepath.Join(meDir, "test05")

	oPath2 := filepath.Join(otherDir, "test02")
	oPath3 := filepath.Join(otherDir, "test03")
	oPath4 := filepath.Join(otherDir, "test04")
	oPath6 := filepath.Join(otherDir, "test06")

	// me
	makeDummyFile(t, mPath1, &alg.Alg)
	makeDummyFile(t, mPath2, &alg.Alg)
	makeDummyFile(t, mPath4, &alg.Alg)
	makeDummyFile(t, mPath5, &alg.Alg)

	// other
	copyFile(t, meDir, otherDir, "test01")
	makeDummyFile(t, oPath2, &alg.Alg)
	touchDelta(t, mPath2, oPath2, time.Minute)
	makeDummyFile(t, oPath3, &alg.Alg)
	copyFile(t, otherDir, meDir, "test03")
	// nolint:errcheck
	os.Rename(filepath.Join(meDir, "test03"), filepath.Join(meDir, "test07"))
	makeDummyFile(t, oPath4, &alg.Alg)
	touchDelta(t, mPath4, oPath4, 0)
	makeDummyFile(t, oPath6, &alg.Alg)

	return meDir, otherDir
}
