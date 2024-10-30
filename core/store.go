package core

import "sync"

type store struct {
	dataStore map[string]string
	sync.RWMutex
}

type ErrNoSuckKey struct {
}

func (e *ErrNoSuckKey) Error() string {
	return "no such key"
}

func newStore() *store {
	return &store{
		dataStore: make(map[string]string),
	}
}

func (s *store) put(key, value string) error {
	s.Lock()
	defer s.Unlock()
	s.dataStore[key] = value

	return nil
}

func (s *store) get(key string) (string, error) {
	s.RLock()
	defer s.RUnlock()

	value, ok := s.dataStore[key]

	if !ok {
		return "", &ErrNoSuckKey{}
	}

	return value, nil
}

func (s *store) delete(key string) error {
	s.Lock()
	defer s.Unlock()
	delete(s.dataStore, key)

	return nil
}
