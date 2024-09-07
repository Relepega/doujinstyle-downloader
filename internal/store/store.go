package store

import (
	"fmt"
	"sync"
)

const (
	ERR_KEY_NOT_IN_MAP = "The requested store key is not set"
)

type storeType map[string]interface{}

type Store struct {
	mu sync.Mutex

	store storeType
}

var store *Store

func GetStore() *Store {
	if store != nil {
		return store
	}

	store = &Store{
		store: make(storeType),
	}

	return store
}

func (s *Store) Set(k string, v interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.store[k] = v
}

func (s *Store) Get(k string) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, ok := s.store[k]
	if !ok {
		return nil, fmt.Errorf(ERR_KEY_NOT_IN_MAP)
	}

	return val, nil
}
