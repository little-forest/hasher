package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/little-forest/hasher/common"
	"github.com/pkg/errors"
)

type DirDiff struct {
	files map[string]*FileDiff
	Path  string
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

func (d DirDiff) GetByStatus(basename string, status DiffStatus) *FileDiff {
	f := d.files[basename]
	if f != nil && f.Status == status {
		return f
	}
	return nil
}

// Find FileDiff that have the same hash value as the target
// and whose status is UNKNOWN.
func (d DirDiff) GetByHash(target *FileDiff) *FileDiff {
	var found *FileDiff

	for basename := range d.files {
		f := d.files[basename]

		if target.CompareHash(f) && f.Status == UNKNOWN {
			if found != nil {
				// There are multiple files with the same hash
				return nil
			}
			found = f
		}
	}
	return found
}

func (d DirDiff) GetChildren() []*FileDiff {
	var children = make([]*FileDiff, len(d.files))
	var i = 0
	for basename := range d.files {
		children[i] = d.files[basename]
		i++
	}
	return children
}

func (d DirDiff) GetSortedChildren() []*FileDiff {
	basenames := make([]string, 0, len(d.files))
	children := make([]*FileDiff, 0, len(d.files))
	for b := range d.files {
		basenames = append(basenames, b)
	}
	sort.Strings(basenames)
	for _, b := range basenames {
		children = append(children, d.files[b])
	}
	return children
}

// Mark given status to all children
func (d *DirDiff) MarkAll(status DiffStatus) {
	for _, f := range d.files {
		f.Status = status
	}
}

// Returns true if all children has SAME status.
func (d DirDiff) IsAllSame() bool {
	for _, f := range d.files {
		if f.Status != SAME {
			return false
		}
	}
	return true
}

func (me *DirDiff) Compare(other *DirDiff) {
	myChildren := me.GetChildren()

	// 1st pass
	for i, mf := range myChildren {
		of := other.GetByStatus(mf.Basename, UNKNOWN)
		if of == nil {
			// 相手がいない
			continue
		}
		// compare and update FileDiff status
		mf.Compare(of)
		myChildren[i] = nil
	}

	// 2nd pass (detect renamed files)
	for i, mf := range myChildren {
		if mf == nil {
			continue
		}

		found := other.GetByHash(mf)
		if found != nil {
			// renamed
			mf.Compare(found)
			myChildren[i] = nil
		}
	}

	// 3rd pass marl added/removed
	for _, mf := range myChildren {
		if mf == nil {
			continue
		}

		mf.Status = ADDED
	}
	for _, of := range other.GetChildren() {
		if of.Status == UNKNOWN {
			of.Status = REMOVED
			me.add(of)
		}
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
			filePath := filepath.Join(dirPath, fileInfo.Name())
			if fileInfo.Type() == fs.ModeSymlink {
				common.ShowWarn("Skip symbolic link %s", filePath)
				continue
			}
			f, err := NewFileDiff(filePath, alg)
			if err != nil {
				common.ShowWarn("Failed to calc hash %s", err.Error())
				continue
			}
			f.Parent = dirDiff
			dirDiff.add(f)
		}
	}

	return dirDiff, nil
}

func DirDiffRecursively(baseDir string, targetDir string) ([]*DirPair, error) {
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
			// TODO:
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}
		dd.MarkAll(ADDED)
		dirPairs = append(dirPairs, NewBaseOnlyDirPair(dd))
	}

	// directories in `targetDir` (removed)
	removedDirList := targetDirList.Difference(baseDirList)
	for p := range removedDirList.Iterator().C {
		dd, err := NewDirDiff(filepath.Join(targetDir, p), alg)
		if err != nil {
			// TODO:
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}
		dd.MarkAll(REMOVED)
		dirPairs = append(dirPairs, NewTargetOnlyDirPair(dd))
	}

	// check intersect directories
	for p := range baseDirList.Intersect(targetDirList).Iterator().C {
		baseDirDiff, err := NewDirDiff(filepath.Join(baseDir, p), alg)
		if err != nil {
			return nil, err
		}
		targetDirDiff, err := NewDirDiff(filepath.Join(targetDir, p), alg)
		if err != nil {
			return nil, err
		}

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
