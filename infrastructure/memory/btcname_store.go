package memory

import (
	"sync"
	"yarr/types"
)

type BtcNameStore struct {
	mu    sync.RWMutex
	store map[string]types.BtcName
}

func NewBtcNameStore() *BtcNameStore {
	return &BtcNameStore{
		store: make(map[string]types.BtcName),
	}
}

func (s *BtcNameStore) Store(btcName *types.BtcName) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.store[btcName.Domain] = *btcName
	return nil
}

func (s *BtcNameStore) Retrieve(key string) (*types.BtcName, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if btcName, exists := s.store[key]; exists {
		return &btcName, nil
	}
	return nil, nil
}

func (s *BtcNameStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.store, key)
	return nil
}
