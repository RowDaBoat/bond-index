package memory

import (
	"sync"
)

type BlockStore struct {
	mu    sync.RWMutex
	block uint64
}

func NewBlockStore() *BlockStore {
	return &BlockStore{}
}

func (s *BlockStore) GetLastIndexedBlock() (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.block, nil
}

func (s *BlockStore) SetLastIndexedBlock(blockId uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.block = blockId
	return nil
}
