package db

import (
	"fmt"
	"path/filepath"

	dbm "github.com/tendermint/tm-db"

	"github.com/bdware/go-datastore/key"
	leveldb "github.com/bdware/go-ds-leveldb"
	dsdb "github.com/bdware/tm-db-go-datastore"
)

// NewDB creates a new database of type backend with the given name.
func NewDB(name string, backend dbm.BackendType, dir string) (dbm.DB, error) {
	if backend == dsdb.GoDatastoreBackend {
		ds, err := leveldb.NewDatastore(filepath.Join(dir, name+".db"), key.KeyTypeBytes, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}
		return dsdb.New(ds), nil
	}
	return dbm.NewDB(name, backend, dir)
}
