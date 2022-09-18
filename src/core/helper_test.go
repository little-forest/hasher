package core

import (
	"crypto"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func assertFileDiff(t *testing.T, expectedBaseName string, expectedStatus DiffStatus, expectedPairName string, actual *FileDiff) {
	t.Helper()
	t.Logf("%s/%s %s %s", actual.Parent.Path, actual.Basename, actual.StatusMark(), actual.PairFileName)
	assert.Equal(t, expectedBaseName, actual.Basename)
	assert.Equal(t, expectedStatus, actual.Status)
	assert.Equal(t, expectedPairName, actual.PairFileName)
}

func copyFile(t *testing.T, srcDir string, dstDir string, filename string) {
	t.Helper()

	src, err := os.Open(filepath.Join(srcDir, filename))
	assert.NoError(t, err)
	//nolint:errcheck
	defer src.Close()

	dst, err := os.Create(filepath.Join(dstDir, filename))
	assert.NoError(t, err)
	//nolint:errcheck
	defer dst.Close()

	_, err = io.Copy(dst, src)
	assert.NoError(t, err)

	// touch timestamp(mtime) same as source file
	srcInfo, _ := src.Stat()
	// nolint:errcheck
	os.Chtimes(dst.Name(), time.Now(), srcInfo.ModTime())
}

// Set the delta duration for mtime of path2 to path1
func touchDelta(t *testing.T, path1 string, path2 string, delta time.Duration) {
	t.Helper()

	stat2, err := os.Stat(path2)
	assert.NoError(t, err)

	// nolint:errcheck
	os.Chtimes(path1, time.Now(), stat2.ModTime().Add(delta))
}

// // Make path1 file's modtime newer than path2 file.
// func touchNewer(t *testing.T, path1 string, path2 string) {
// 	t.Helper()
//
// 	stat2, err := os.Stat(path2)
// 	assert.NoError(t, err)
//
// 	os.Chtimes(path1, time.Now(), stat2.ModTime().Add(time.Minute))
// }
//
// // Make path1 file's modtime older than path2 file.
// func touchOlder(t *testing.T, path1 string, path2 string) {
// 	t.Helper()
//
// 	stat2, err := os.Stat(path2)
// 	assert.NoError(t, err)
//
// 	os.Chtimes(path1, time.Now(), stat2.ModTime().Sub(time.Minute))
// }

// Make dummy file and returns it's hash value.
func makeDummyFile(t *testing.T, path string, alg *crypto.Hash) string {
	const SIZE = 128

	t.Helper()

	rand := rand.New(rand.NewSource(time.Now().UnixNano()))

	file, err := os.Create(path)
	assert.NoError(t, err)
	//nolint:errcheck
	defer file.Close()

	// create random bytes contents
	var contents [SIZE]byte
	for i := 0; i < SIZE; i++ {
		contents[i] = byte(rand.Int() & 0xff)
	}
	contentsStr := fmt.Sprintf("%x", contents)

	//nolint:errcheck
	file.WriteString(contentsStr)

	// calc hash
	hash := alg.New()
	hash.Write([]byte(contentsStr))
	hashString := fmt.Sprintf("%x", hash.Sum(nil))
	return hashString
}

func makeSingleDummyFile(t *testing.T, alg *crypto.Hash) (string, string) {
	t.Helper()

	tmpDir := t.TempDir()

	path := filepath.Join(tmpDir, randomString(10))
	hash := makeDummyFile(t, path, alg)

	fmt.Printf("dummy: %s\n", path)
	return path, hash
}

func randomString(n int) string {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))

	var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
