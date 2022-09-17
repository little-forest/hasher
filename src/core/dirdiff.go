package core

import (
	"fmt"
	"os"
	"path/filepath"
)

type DirDiff struct {
	Path  string
	files map[string]*FileDiff
}

func (d *DirDiff) add(f *FileDiff) {
	d.files[f.Basename] = f
}

func (d DirDiff) Get(fileName string) *FileDiff {
	return d.files[fileName]
}

func (d DirDiff) Count() int {
	return len(d.files)
}

func (me *DirDiff) Compare(other *DirDiff) {
	for basename := range me.files {
		fmt.Printf("bn : %s\n", basename)
	}
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
