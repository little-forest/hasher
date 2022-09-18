package core

import (
	"fmt"
	"io"
	"strconv"

	. "github.com/little-forest/hasher/common"
)

const hashBufSize = 256 * 1024

const Xattr_prefix = "user.hasher"
const Xattr_size = Xattr_prefix + ".size"
const Xattr_modifiedTime = Xattr_prefix + ".mtime"

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
		if curSize := GetXattr(file, Xattr_size); size != curSize {
			changed = true
		} else if curMtime := GetXattr(file, Xattr_modifiedTime); modTime != curMtime {
			changed = true
		}
		if !forceUpdate && !changed {
			return false, curHash, nil
		}
	}

	hash, err := calcHashString(file, alg)
	if err != nil {
		return false, "", err
	}

	if err := SetXattr(file, alg.AttrName, hash); err != nil {
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
