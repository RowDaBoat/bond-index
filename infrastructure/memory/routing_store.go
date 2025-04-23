package memory

import (
	"sync"
	"yarr/types"
)

type RoutingStore struct {
	mu    sync.RWMutex
	store map[string]types.Routing
}

func NewRoutingStore() *RoutingStore {
	return &RoutingStore{
		store: make(map[string]types.Routing),
	}
}

func (s *RoutingStore) Store(address string, name string, value *types.Routing) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := address + ":" + name
	s.store[key] = *value
	return nil
}

func (s *RoutingStore) Retrieve(address string, name string) (*types.Routing, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := address + ":" + name
	if routing, exists := s.store[key]; exists {
		return &routing, nil
	}
	return nil, nil
}

func (s *RoutingStore) Delete(address string, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := address + ":" + name
	delete(s.store, key)
	return nil
}
