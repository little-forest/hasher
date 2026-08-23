package hashcore

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

const hashBufSize = 256 * 1024

const Xattr_prefix = "user.hasher"

// File size when hash is updated
const Xattr_size = Xattr_prefix + ".size"

// File modification time when hash is updated
const Xattr_modifiedTime = Xattr_prefix + ".mtime"

// Time of hash update
const Xattr_hashCheckedTime = Xattr_prefix + ".htime"

// UpdateHashStictly updates specified file's hash value.
// Returns an UpdateError if the update of an attribute fails.
//
//	changed : bool
//	hash value : *Hash
//	error : error
func UpdateHashStrictly(path string, alg *HashAlg, forceUpdate bool) (bool, *Hash, error) {
	file, err := OpenFile(path)
	if err != nil {
		return false, nil, err
	}
	// nolint:errcheck
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return false, nil, err
	}
	size := fmt.Sprint(info.Size())
	modTime := strconv.FormatInt(info.ModTime().UnixNano(), 10)

	var changed bool
	curHash := GetXattr(file, alg.AttrName)
	if curHash != "" {
		// check if existing hash value is valid
		// If the file size and modtime have not changed, it is considered correct.
		if curSize := GetXattr(file, Xattr_size); size != curSize {
			changed = true
		} else if curMtime := GetXattr(file, Xattr_modifiedTime); modTime != curMtime {
			changed = true
		}
		if !forceUpdate && !changed {
			// update only checked time
			var updateErr error
			if e := updateHashCheckedTime(file); e != nil {
				updateErr = NewUpdateError(e)
			}
			hash, _ := NewHashFromString(path, alg, curHash, info.ModTime().Unix())
			return false, hash, updateErr
		}
	}

	// do calculate hash value
	hash, err := CalcHash(path, alg)
	if err != nil {
		return false, nil, err
	}

	// update attributes
	if err := SetXattr(file, alg.AttrName, hash.String()); err != nil {
		return true, hash, NewUpdateError(err)
	}
	if err := updateHashCheckedTime(file); err != nil {
		return true, hash, NewUpdateError(err)
	}
	if err := SetXattr(file, Xattr_size, size); err != nil {
		return true, hash, NewUpdateError(err)
	}
	if err := SetXattr(file, Xattr_modifiedTime, modTime); err != nil {
		return true, hash, NewUpdateError(err)
	}

	return true, hash, nil
}

func updateHashCheckedTime(f *os.File) error {
	htime := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	if err := SetXattr(f, Xattr_hashCheckedTime, htime); err != nil {
		return err
	}
	return nil
}

func CalcHash(path string, hashAlg *HashAlg) (*Hash, error) {
	if !hashAlg.Alg.Available() {
		return nil, fmt.Errorf("no implementation")
	}

	r, err := OpenFile(path)
	if err != nil {
		return nil, err
	}

	hash := hashAlg.Alg.New()
	if _, err := io.CopyBuffer(hash, r, make([]byte, hashBufSize)); err != nil {
		return nil, err
	}

	info, _ := os.Stat(path)

	return NewHash(path, hashAlg, hash.Sum(nil), info.ModTime().Unix()), nil
}

// Get hash value.
// This function will not check hash is updated.
// When given file's hash has not been calculated, it will return nil.
func GetHash(path string, alg *HashAlg) (*Hash, error) {
	file, err := OpenFile(path)
	if err != nil {
		return nil, err
	}
	// nolint:errcheck
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	curHash := GetXattr(file, alg.AttrName)
	if curHash != "" {
		hash, _ := NewHashFromString(path, alg, curHash, info.ModTime().Unix())
		return hash, nil
	} else {
		return nil, nil
	}
}
