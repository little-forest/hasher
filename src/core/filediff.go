package core

import (
	"os"
	"path/filepath"
	"time"
)

type DiffStatus uint8

const (
	UNKNOWN DiffStatus = iota + 1
	ADDED
	SAME
	NOT_SAME_NEW
	NOT_SAME_OLD
	NOT_SAME
	RENAMED
	REMOVED
)

type FileDiff struct {
	Basename     string
	PairFileName string
	Parent       *DirDiff
	HashValue    []byte
	ModTime      time.Time
	Status       DiffStatus
}

func NewFileDiff(filePath string, alg *HashAlg) (*FileDiff, error) {
	_, hash, err := UpdateHash2(filePath, alg, false)
	if err != nil {
		return nil, err
	}

	basename := filepath.Base(filePath)

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	modtime := info.ModTime()

	d := &FileDiff{
		Basename:     basename,
		PairFileName: "",
		HashValue:    hash.Value,
		ModTime:      modtime,
		Status:       UNKNOWN,
	}

	return d, nil
}

// Compare only each other's hash value only.
func (me FileDiff) CompareHash(other *FileDiff) bool {
	return arrayEquals(me.HashValue, other.HashValue)
}

// Compare each other and update FileDiff status.
func (me *FileDiff) Compare(other *FileDiff) bool {
	me.PairFileName = other.Basename
	other.PairFileName = me.Basename

	if arrayEquals(me.HashValue, other.HashValue) {
		// same
		if me.PairFileName == other.PairFileName {
			me.Status = SAME
			other.Status = SAME
		} else {
			me.Status = RENAMED
			other.Status = RENAMED
		}
		return true
	}

	// not same
	if me.ModTime.After(other.ModTime) {
		me.Status = NOT_SAME_NEW
		other.Status = NOT_SAME_OLD
	} else if me.ModTime.Before(other.ModTime) {
		me.Status = NOT_SAME_OLD
		other.Status = NOT_SAME_NEW
	} else {
		me.Status = NOT_SAME
		other.Status = NOT_SAME
	}
	return false
}

func arrayEquals(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func (f *FileDiff) StatusMark() string {
	switch f.Status {
	case UNKNOWN:
		return "[?]"
	case ADDED:
		return "[+]"
	case SAME:
		return "[=]"
	case NOT_SAME_NEW:
		return "[>]"
	case NOT_SAME_OLD:
		return "[<]"
	case NOT_SAME:
		return "[~]"
	case RENAMED:
		return "[R]"
	case REMOVED:
		return "[-]"
	}
	return ""
}
