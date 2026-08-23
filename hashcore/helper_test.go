package hashcore

import (
	"crypto"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
