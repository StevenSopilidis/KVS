package db

import "sync"

type Store struct {
	dataStore map[string]string
	sync.RWMutex
}

type ErrNoSuckKey struct {
}

func (e *ErrNoSuckKey) Error() string {
	return "no such key"
}

func NewStore() *Store {
	return &Store{
		dataStore: make(map[string]string),
	}
}

func (s *Store) Put(key, value string) error {
	s.Lock()
	defer s.Unlock()
	s.dataStore[key] = value

	return nil
}

func (s *Store) Get(key string) (string, error) {
	s.RLock()
	defer s.RUnlock()

	value, ok := s.dataStore[key]

	if !ok {
		return "", &ErrNoSuckKey{}
	}

	return value, nil
}

func (s *Store) Delete(key string) error {
	s.Lock()
	defer s.Unlock()
	delete(s.dataStore, key)

	return nil
}
