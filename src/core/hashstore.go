package core

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type HashStore struct {
	store map[string][]*Hash
	size  int
}

func NewHashStore() *HashStore {
	return &HashStore{store: make(map[string][]*Hash)}
}

func (s *HashStore) Put(hash *Hash) {
	key := hash.String()

	if s.store[key] == nil {
		s.store[key] = []*Hash{hash}
	} else {
		s.store[key] = append(s.store[key], hash)
	}

	s.size++
}

func (s HashStore) Get(hashValue string) []*Hash {
	return s.store[hashValue]
}

func (s HashStore) KeySet() []string {
	keys := make([]string, len(s.store))
	idx := 0
	for k := range s.store {
		keys[idx] = k
		idx++
	}
	return keys
}

func (s HashStore) Size() int {
	return s.size
}

func (s HashStore) Values() []*Hash {
	values := make([]*Hash, s.size)

	idx := 0
	for k := range s.store {
		for _, h := range s.store[k] {
			values[idx] = h
			idx++
		}
	}

	sort.Slice(values, func(i, j int) bool {
		return strings.Compare(values[i].Path, values[j].Path) < 0
	})

	return values
}

func (s *HashStore) LoadHashData(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = '\t'
	r.Comment = '#'
	r.LazyQuotes = true

	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "%s : %s : %v\n", path, err.Error(), line)
			continue
		}

		hash, err := s.parseHashLine(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		}

		s.Put(hash)
	}

	return nil
}

func (s HashStore) parseHashLine(line []string) (*Hash, error) {
	if len(line) < 4 {
		return nil, fmt.Errorf("invalid format : %v", line)
	}
	modTime, err := strconv.Atoi(line[2])
	if err != nil {
		return nil, fmt.Errorf("failed to parse modTime : %v", line)
	}

	pos := strings.Index(line[3], ":")
	if pos == -1 {
		return nil, fmt.Errorf("invalid hash value format : %v", line)
	}
	alg := NewHashAlgFromString(line[3][0:pos])
	hashValue := line[3][pos+1:]

	hash, err := NewHashFromString(line[0], alg, hashValue, int64(modTime))
	if err != nil {
		return nil, fmt.Errorf("failed to patse tsv : %v", line)
	}
	return hash, nil
}

func (s *HashStore) AppendHashDataFromDirectory(dirPath string, alg *HashAlg, verbose bool) error {
	err := filepath.WalkDir(dirPath, func(path string, info fs.DirEntry, e error) error {
		if e != nil {
			return errors.Wrap(e, "failed to filepath.Walk")
		}

		if info.IsDir() {
			return nil
		}

		if verbose {
			// TODO: verbose
			// fmt.Printf("\x1b7\x1b[0J%d/%d %s\x1b8", count, totalCount, path)
			fmt.Printf("verbose\n")
		}
		absPath, _ := filepath.Abs(path)
		_, hash, e := UpdateHash(absPath, alg, false)
		if e != nil {
			fmt.Fprintf(os.Stderr, "Failed to update hash : %s (reason : %s)\n", absPath, e.Error())
		} else {
			s.Put(hash)
		}
		return nil
	})
	return err
}
