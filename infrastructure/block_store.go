package infrastructure

import (
	"encoding/binary"
	"fmt"
	"yarr/service"

	"github.com/dgraph-io/badger/v4"
)

var _ service.BlockStore = (*BlockStore)(nil)

type BlockStore struct {
	store *Store
}

func NewBlockStore(store *Store) *BlockStore {
	return &BlockStore{store: store}
}

func (s *BlockStore) GetLastIndexedBlock() (uint64, error) {
	var value []byte
	err := s.store.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("last_indexed_block"))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			value = val
			return nil
		})
	})

	if err != nil {
		return 0, fmt.Errorf("failed to get last indexed block: %w", err)
	}

	if len(value) == 0 {
		return 0, nil
	}

	return binary.BigEndian.Uint64(value), nil
}

func (s *BlockStore) SetLastIndexedBlock(blockId uint64) error {
	value := make([]byte, 8)
	binary.BigEndian.PutUint64(value, blockId)

	return s.store.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("last_indexed_block"), value)
	})
}
