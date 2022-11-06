package core

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
)

// ------------------------------------------------------------------------------
//  hash file format
//
//  1: full path
//  2: file name
//  3: file modified timestamp (UNIX time)
//  4: sha1 hash
// ===============================================================================

type Hash struct {
	Path    string
	Alg     *HashAlg
	Value   []byte
	ModTime int64 // unix time
}

func NewHash(path string, alg *HashAlg, value []byte, modTime int64) *Hash {
	return &Hash{
		Path:    path,
		Alg:     alg,
		Value:   value,
		ModTime: modTime,
	}
}

func NewHashFromString(path string, alg *HashAlg, value string, modTime int64) (*Hash, error) {
	bytes, err := hex.DecodeString(value)
	if err != nil {
		return nil, err
	}

	return &Hash{
		Path:    path,
		Alg:     alg,
		Value:   bytes,
		ModTime: modTime,
	}, nil
}

func (h Hash) String() string {
	return fmt.Sprintf("%x", h.Value)
}

func (h Hash) Json() string {
	return fmt.Sprintf("{\"path\": \"%s\", \"hash\": \"%s:%s\"}", h.Path, h.Alg.AlgName, h.String())
}

func (h Hash) DollyTsv() string {
	basename := filepath.Base(h.Path)
	return fmt.Sprintf("%s\t%s\t%d\t%s:%s", h.Path, basename, h.ModTime, h.Alg.AlgName, h.String())
}

func (h Hash) HasSameHashValue(other *Hash) bool {
	if other == nil {
		return false
	}
	if len(h.Value) != len(other.Value) {
		return false
	}

	for i, b := range h.Value {
		if b != other.Value[i] {
			return false
		}
	}
	return true
}
