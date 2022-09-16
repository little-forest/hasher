package core

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
)

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

func loadHashDataToByteArray(t *testing.T, basename string, algName string) []byte {
	hash := loadHashData(t, basename, algName)
	bytes, _ := hex.DecodeString(hash)
	return bytes
}
