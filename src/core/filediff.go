package core

import (
	"encoding/hex"
	"path/filepath"
)

type DiffStatus uint8

const (
	UNKNOWN DiffStatus = iota + 1
	ADDED
	NOT_SAME_NEW
	NOT_SAME_OLD
	NOT_SAME
	REMOVED
)

type FileDiff struct {
	Basename     string
	PairFileName string
	HashValue    []byte
	Status       DiffStatus
}

func NewFileDiff(filePath string, alg *HashAlg) (*FileDiff, error) {
	_, hash, err := UpdateHash(filePath, alg, false)
	if err != nil {
		return nil, err
	}

	hashBytes, _ := hex.DecodeString(hash)
	basename := filepath.Base(filePath)

	d := &FileDiff{
		Basename:     basename,
		PairFileName: "",
		HashValue:    hashBytes,
		Status:       UNKNOWN,
	}

	return d, nil
}
