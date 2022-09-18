package core

import (
	"os"
	"path/filepath"
	"sort"
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
			// TODO symbolic link check
			filePath := filepath.Join(dirPath, fileInfo.Name())
			f, err := NewFileDiff(filePath, alg)
			if err != nil {
				return nil, err
			}
			f.Parent = dirDiff
			dirDiff.add(f)
		}
	}

	return dirDiff, nil
}
