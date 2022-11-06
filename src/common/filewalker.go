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

func WalkDirs(dirPaths []string, walker FileWalker) error {
	// check dirctories
	for _, path := range dirPaths {
		err := EnsureDirectory(path)
		if err != nil {
			return err
		}
	}

	for _, dirPath := range dirPaths {
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

			walkError := walker.Deal(f)
			return walkError
		})
		if err != nil {
			return err
		}
	}
	return nil
}
