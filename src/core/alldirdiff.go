package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/pkg/errors"
)

type DirPairStatus uint8

const (
	BASE_ONLY = iota + 1
	PAIR
	TARGET_ONLY
)

type DirPair struct {
	Base   *DirDiff
	Target *DirDiff
	Status DirPairStatus
}

func NewDirPair(base *DirDiff, target *DirDiff) *DirPair {
	return &DirPair{
		Base:   base,
		Target: target,
		Status: PAIR,
	}
}

func NewBaseOnlyDirPair(base *DirDiff) *DirPair {
	return &DirPair{
		Base:   base,
		Status: BASE_ONLY,
	}
}

func NewTargetOnlyDirPair(target *DirDiff) *DirPair {
	return &DirPair{
		Target: target,
		Status: TARGET_ONLY,
	}
}

func (d DirPair) Path() string {
	if d.Base != nil {
		return d.Base.Path
	} else if d.Target != nil {
		return d.Target.Path
	} else {
		return ""
	}
}

func DirDiffRecursive(baseDir string, targetDir string) ([]*DirPair, error) {
	alg := NewDefaultHashAlg()

	// list directories
	baseDir = normalizeDirPath(baseDir)
	baseDirList, err := listDirectories(baseDir)
	if err != nil {
		return nil, err
	}
	targetDir = normalizeDirPath(targetDir)
	targetDirList, err := listDirectories(targetDir)
	if err != nil {
		return nil, err
	}

	var dirPairs []*DirPair

	// directories in `baseDir` (added)
	baseonly := baseDirList.Difference(targetDirList)
	for p := range baseonly.Iterator().C {
		dd, err := NewDirDiff(filepath.Join(baseDir, p), alg)
		if err != nil {
			// TODO
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}
		dirPairs = append(dirPairs, NewBaseOnlyDirPair(dd))
	}

	// directories in `targetDir` (removed)
	removedDirList := targetDirList.Difference(baseDirList)
	for p := range removedDirList.Iterator().C {
		dd, err := NewDirDiff(filepath.Join(targetDir, p), alg)
		if err != nil {
			// TODO
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}
		dirPairs = append(dirPairs, NewTargetOnlyDirPair(dd))
	}

	// check intersect directories
	for p := range baseDirList.Intersect(targetDirList).Iterator().C {
		baseDirDiff, _ := NewDirDiff(filepath.Join(baseDir, p), alg)
		targetDirDiff, _ := NewDirDiff(filepath.Join(targetDir, p), alg)

		baseDirDiff.Compare(targetDirDiff)
		dirPairs = append(dirPairs, NewDirPair(baseDirDiff, targetDirDiff))
	}

	sort.Slice(dirPairs[:], func(i, j int) bool {
		return strings.Compare(dirPairs[i].Path(), dirPairs[j].Path()) < 0
	})

	return dirPairs, nil
}

func normalizeDirPath(dirpath string) string {
	dirpath = filepath.Clean(dirpath)
	if !strings.HasSuffix(dirpath, "/") {
		dirpath = dirpath + "/"
	}
	return dirpath
}

func listDirectories(dir string) (mapset.Set[string], error) {
	dirlist := mapset.NewSet[string]()

	err := filepath.WalkDir(dir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrap(err, "failed to filepath.Walk")
		}

		if info.IsDir() {
			dirpath := strings.TrimPrefix(path, dir)
			if dirpath != "" {
				dirlist.Add(dirpath)
			}
		}
		return nil
	})
	return dirlist, err
}
