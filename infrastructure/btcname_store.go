package infrastructure

import (
	"encoding/json"
	"fmt"
	"yarr/service"
	"yarr/types"

	"github.com/dgraph-io/badger/v4"
)

var _ service.BtcNameStore = (*BtcNameStore)(nil)

type BtcNameStore struct {
	store *Store
}

func NewBtcNameStore(store *Store) *BtcNameStore {
	return &BtcNameStore{store: store}
}

func (s *BtcNameStore) Store(btcName *types.BtcName) error {
	value, _ := json.Marshal(btcName)
	return s.store.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(btcName.Domain), value)
	})
}

func (s *BtcNameStore) Retrieve(btcName string) (*types.BtcName, error) {
	var value []byte
	err := s.store.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(btcName))
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
		return nil, err
	}

	if len(value) == 0 {
		return nil, nil
	}

	var result types.BtcName
	if err := json.Unmarshal(value, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal btcName: %w", err)
	}

	return &result, nil
}

func (s *BtcNameStore) Delete(btcName string) error {
	return s.store.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(btcName))
	})
}
