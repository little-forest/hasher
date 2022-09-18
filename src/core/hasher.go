package core

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	. "github.com/little-forest/hasher/common"
)

const hashBufSize = 256 * 1024

const Xattr_prefix = "user.hasher"

// File size when hash is updated
const Xattr_size = Xattr_prefix + ".size"

// File modification time when hash is updated
const Xattr_modifiedTime = Xattr_prefix + ".mtime"

// Time of hash update
const Xattr_hashCheckedTime = Xattr_prefix + ".htime"

// Update specified file's hash value
//
//	changed : bool
//	hash value : string
//	error : error
func UpdateHash(path string, alg *HashAlg, forceUpdate bool) (bool, string, error) {
	file, err := OpenFile(path)
	if err != nil {
		return false, "", err
	}
	// nolint:errcheck
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return false, "", err
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
			err := updateHashCheckedTime(file)
			return false, curHash, err
		}
	}

	// do calculate hash value
	hash, err := calcHashString(file, alg)
	if err != nil {
		return false, "", err
	}

	// update attributes
	if err := SetXattr(file, alg.AttrName, hash); err != nil {
		return true, hash, err
	}
	if err := updateHashCheckedTime(file); err != nil {
		return true, hash, err
	}
	if err := SetXattr(file, Xattr_size, size); err != nil {
		return true, hash, err
	}
	if err := SetXattr(file, Xattr_modifiedTime, modTime); err != nil {
		return true, hash, err
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

func CalcFileHash(path string, alg *HashAlg) (string, error) {
	r, err := OpenFile(path)
	// nolint:staticcheck,errcheck
	defer r.Close()
	if err != nil {
		return "", err
	}

	hash, err := calcHashString(r, alg)
	return hash, err
}

func calcHashString(r io.Reader, hashAlg *HashAlg) (string, error) {
	if !hashAlg.Alg.Available() {
		return "", fmt.Errorf("no implementation")
	}

	hash := hashAlg.Alg.New()
	if _, err := io.CopyBuffer(hash, r, make([]byte, hashBufSize)); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
