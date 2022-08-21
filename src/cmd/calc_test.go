package cmd

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcFileHash(t *testing.T) {
	basename := "testdata01"
	hashAlg := NewDefaultHashAlg("")

	hash, err := calcFileHash(fmt.Sprintf("testdata/%s.txt", basename), hashAlg)

	expectedHash := loadHashData(t, basename, hashAlg.AlgName)
	assert.NoError(t, err)
	assert.Equal(t, expectedHash, hash)
}

func TestCalcFileHash_failed(t *testing.T) {
	hashAlg := NewDefaultHashAlg("")

	_, err := calcFileHash("dummy", hashAlg)
	assert.Error(t, err)
}

func TestCalcHashString(t *testing.T) {
	basename := "testdata01"
	hashAlg := NewDefaultHashAlg("")

	f := openTestData(t, basename)
	//nolint:errcheck
	defer f.Close()

	hash, err := calcHashString(f, hashAlg)

	expectedHash := loadHashData(t, basename, hashAlg.AlgName)
	assert.NoError(t, err)
	assert.Equal(t, expectedHash, hash)
}

func openTestData(t *testing.T, basename string) *os.File {
	f, err := os.Open(fmt.Sprintf("testdata/%s.txt", basename))
	if err != nil {
		t.Errorf("failed to open. : %v", err)
	}
	return f
}

func loadHashData(t *testing.T, basename string, algName string) string {
	f, err := os.Open(fmt.Sprintf("testdata/%s.%s", basename, algName))
	if err != nil {
		t.Errorf("%v", f)
	}
	//nolint:errcheck
	defer f.Close()

	hash := ""
	s := bufio.NewScanner(f)

	if s.Scan() {
		// read first line
		hash = s.Text()
	}

	return hash
}
