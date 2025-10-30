package infrastructure

import (
	"bond/service"
	"bond/types"
	"bytes"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

var _ service.RoutingStore = (*RoutingStore)(nil)

type RoutingStore struct {
	store *Store
}

func NewRoutingStore(store *Store) *RoutingStore {
	return &RoutingStore{store: store}
}

func (s *RoutingStore) Store(address string, name string, value *types.Routing) error {
	key := []byte(fmt.Sprintf("routing:%s:%s", address, name))

	valueBytes := make([]byte, 0)
	valueBytes = append(valueBytes, []byte(value.Address)...)
	valueBytes = append(valueBytes, 0)
	valueBytes = append(valueBytes, []byte(value.NostrNpub)...)
	valueBytes = append(valueBytes, 0)
	relaysJoined := strings.Join(value.NostrRelays, ",")
	valueBytes = append(valueBytes, []byte(relaysJoined)...)

	return s.store.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, valueBytes)
	})
}

func (s *RoutingStore) Retrieve(address string, name string) (*types.Routing, error) {
	key := []byte(fmt.Sprintf("routing:%s:%s", address, name))
	var value []byte

	err := s.store.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			value = val
			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get routing: %w", err)
	}

	parts := bytes.Split(value, []byte{0})
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid routing data format")
	}

	relaysStr := string(parts[2])

	var relays = strings.Split(relaysStr, ",")

	return &types.Routing{
		Address:     string(parts[0]),
		Domain:      name,
		NostrNpub:   string(parts[1]),
		NostrRelays: relays,
	}, nil
}

func (s *RoutingStore) Delete(address string, name string) error {
	key := []byte(fmt.Sprintf("routing:%s:%s", address, name))
	return s.store.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (s *RoutingStore) Has(address string, name string) (bool, error) {
	key := []byte(fmt.Sprintf("routing:%s:%s", address, name))
	var exists bool

	err := s.store.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			exists = false
			return nil
		}
		if err != nil {
			return err
		}
		exists = true
		return nil
	})

	return exists, err
}
