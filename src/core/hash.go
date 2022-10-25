package core

import (
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

func LoadHashData(path string) (*HashStore, error) {
	store := NewHashStore()

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	// nolint:errcheck
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = '\t'
	r.Comment = '#'

	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse TSV : %s\n", err.Error())
			continue
		}

		hash, err := parseHashLine(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		}

		store.Put(hash)
	}

	return store, nil
}

func parseHashLine(line []string) (*Hash, error) {
	if len(line) < 4 {
		return nil, fmt.Errorf("Invalid format : %v", line)
	}
	modTime, err := strconv.Atoi(line[2])
	if err != nil {
		return nil, fmt.Errorf("Failed to parse modTime : %v", line)
	}

	pos := strings.Index(line[3], ":")
	if pos == -1 {
		return nil, fmt.Errorf("Invalid hash value format : %v", line)
	}
	alg := NewHashAlgFromString(line[3][0:pos])
	hashValue := line[3][pos+1:]

	hash, err := NewHashFromString(line[0], alg, hashValue, int64(modTime))
	if err != nil {
		return nil, fmt.Errorf("Failed to patse tsv : %v", line)
	}
	return hash, nil
}
