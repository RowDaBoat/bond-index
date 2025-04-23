package infrastructure

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
)

type Store struct {
	db *badger.DB
}

func NewStore(path string) *Store {
	expandedPath := os.ExpandEnv(path)
	opts := badger.DefaultOptions(filepath.Join(expandedPath, "btcnames"))
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		fmt.Printf("Error: failed to open database: %v\n", err)
		os.Exit(1)
	}

	return &Store{db: db}
}

func (store *Store) Close() error {
	return store.db.Close()
}
