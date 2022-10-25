package core

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
		}
		idx++
	}

	return values
}
