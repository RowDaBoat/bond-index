package infrastructure

import (
	"bond/service"
	"bond/types"
	"encoding/json"
	"fmt"

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
			fmt.Printf("Retrieve View: key not found: %s\n", btcName)
			return nil
		}
		if err != nil {
			fmt.Printf("Retrieve View: error getting key %s: %v\n", btcName, err)
			return err
		}
		fmt.Printf("Retrieve View: key found, reading value for: %s\n", btcName)
		return item.Value(func(val []byte) error {
			value = make([]byte, len(val))
			copy(value, val)
			fmt.Printf("Retrieve View: copied %d bytes for: %s\n", len(val), btcName)
			return nil
		})
	})

	if err != nil {
		fmt.Printf("Retrieve: returning error from transaction: %v\n", err)
		return nil, err
	}

	if len(value) == 0 {
		fmt.Printf("Retrieve: value is empty (key not found) for: %s\n", btcName)
		return nil, nil
	}

	var result types.BtcName
	if err := json.Unmarshal(value, &result); err != nil {
		fmt.Printf("Retrieve: unmarshal error for %s: %v\n", btcName, err)
		return nil, fmt.Errorf("failed to unmarshal btcName: %w", err)
	}

	fmt.Printf("Retrieve: successfully retrieved %s\n", btcName)
	return &result, nil
}

func (s *BtcNameStore) Delete(btcName string) error {
	return s.store.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(btcName))
	})
}
