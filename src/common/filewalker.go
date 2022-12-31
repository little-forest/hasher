package common

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type FileWalker interface {
	Deal(file *os.File) error
}

func WalkDirsWithWalker(dirPaths []string, walker FileWalker) error {
	return WalkDirs(dirPaths, func(f *os.File) error {
		return walker.Deal(f)
	})
}

func WalkDir(dirPath string, dealFile func(file *os.File) error) error {
	err := filepath.WalkDir(dirPath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrap(err, "failed to filepath.Walk")
		}

		if info.IsDir() {
			return nil
		}

		// don't follow symbolic link
		isSymlink, err := IsSymbolicLink(path)
		if err != nil {
			return err
		}
		if isSymlink {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		// nolint:errcheck
		defer f.Close()

		walkError := dealFile(f)
		return walkError
	})
	return err
}

func WalkDirs(dirPaths []string, dealFile func(file *os.File) error) error {
	// check dirctories
	for _, path := range dirPaths {
		err := EnsureDirectory(path)
		if err != nil {
			return err
		}
	}

	for _, dirPath := range dirPaths {
		err := WalkDir(dirPath, dealFile)
		if err != nil {
			return err
		}
	}
	return nil
}
