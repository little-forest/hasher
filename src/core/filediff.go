package core

import "encoding/hex"

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
	FileName     string
	PairFileName string
	HashValue    []byte
	Status       DiffStatus
}

type DirDiff struct {
	Path  string
	Files map[string]*FileDiff
}

func NewFileDiff(filePath string, alg *HashAlg) (*FileDiff, error) {
	_, hash, err := UpdateHash(filePath, alg, false)
	if err != nil {
		return nil, err
	}

	hashBytes, _ := hex.DecodeString(hash)

	d := &FileDiff{
		FileName:     filePath,
		PairFileName: "",
		HashValue:    hashBytes,
		Status:       UNKNOWN,
	}

	return d, nil
}
