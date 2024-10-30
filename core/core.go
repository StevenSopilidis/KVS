package core

import (
	"github.com/StevenSopilidis/kvs/persistance"
)

// implements core business logic
type KeyValueStore struct {
	store  *store
	logger persistance.TransactionLogger
}

func NewKeyValueStore(logger persistance.TransactionLogger) (*KeyValueStore, error) {
	var err error

	store := newStore()

	// read and replay all past events from log file
	events, errors := logger.ReadEvents()
	e, ok := persistance.Event{}, true

	for ok && err == nil {
		select {
		case err, ok = <-errors:
		case e, ok = <-events:
			switch e.Type {
			case persistance.EventPut:
				store.put(e.Key, e.Value)
			case persistance.EventDelete:
				store.delete(e.Key)
			}
		}
	}

	if err != nil {
		return nil, err
	}

	logger.Run()

	return &KeyValueStore{
		store:  newStore(),
		logger: logger,
	}, nil
}

type PutRequest struct {
	Data string
}

func (s *KeyValueStore) Put(key, value string) error {
	err := s.store.put(key, value)

	if err != nil {
		return err
	}

	s.logger.WritePut(key, value)

	return nil
}

type GetResponse struct {
	Value string
}

func (s *KeyValueStore) Get(key string) (string, error) {
	value, err := s.store.get(key)

	if err != nil {
		return "", err
	}

	return value, err
}

func (s *KeyValueStore) Delete(key string) error {
	err := s.store.delete(key)

	return err
}
