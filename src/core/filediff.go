package core

import (
	"encoding/hex"
	"os"
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

type DirDiff struct {
	Path  string
	files map[string]*FileDiff
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

func (d *DirDiff) add(f *FileDiff) {
	d.files[f.Basename] = f
}

func (d DirDiff) Get(fileName string) *FileDiff {
	return d.files[fileName]
}

func NewDirDiff(dirPath string, alg *HashAlg) (*DirDiff, error) {
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}

	fileInfos, err := dir.ReadDir(-1)
	if err != nil {
		return nil, err
	}

	dirDiff := &DirDiff{
		Path:  dirPath,
		files: make(map[string]*FileDiff),
	}

	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() {
			// TODO symbolic link check
			filePath := filepath.Join(dirPath, fileInfo.Name())
			f, err := NewFileDiff(filePath, alg)
			if err != nil {
				return nil, err
			}
			dirDiff.add(f)
		}
	}

	return dirDiff, nil
}
