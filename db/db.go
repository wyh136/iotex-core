// Copyright (c) 2018 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package db

import (
	"sync"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"

	"github.com/iotexproject/iotex-core/common/service"
)

var (
	// ErrInvalidDB indicates invalid operation attempted to Blockchain database
	ErrInvalidDB = errors.New("invalid DB operation")
	// ErrNotExist indicates certain item does not exist in Blockchain database
	ErrNotExist = errors.New("not exist in DB")
	// ErrAlreadyExist indicates certain item already exists in Blockchain database
	ErrAlreadyExist = errors.New("already exist in DB")
)

// KVStore is the interface of KV store.
type KVStore interface {
	service.Service
	// Put insert or update a record identified by (namespace, key)
	Put(string, []byte, []byte) error
	// BatchPut insert or update a slice of records identified by (namespace, key)
	BatchPut(string, [][]byte, [][]byte) error
	// Put puts a record only if (namespace, key) doesn't exist, otherwise return ErrAlreadyExist
	PutIfNotExists(string, []byte, []byte) error
	// Get gets a record by (namespace, key)
	Get(string, []byte) ([]byte, error)
	// Delete deletes a record by (namespace, key)
	Delete(string, []byte) error
}

const (
	keyDelimiter = "."
)

// memKVStore is the in-memory implementation of KVStore for testing purpose
type memKVStore struct {
	service.AbstractService
	data sync.Map
}

// NewMemKVStore instantiates an in-memory KV store
func NewMemKVStore() KVStore {
	return &memKVStore{}
}

// Put inserts a <key, value> record
func (m *memKVStore) Put(namespace string, key []byte, value []byte) error {
	m.data.Store(namespace+keyDelimiter+string(key), value)
	return nil
}

// BatchPut inserts a slice of records <key[], value[]>
func (m *memKVStore) BatchPut(namespace string, key [][]byte, value [][]byte) error {
	if len(key) != len(value) {
		return errors.Wrap(ErrInvalidDB, "batch put <k, v> size not match")
	}
	for i := 0; i < len(key); i++ {
		m.data.Store(namespace+keyDelimiter+string(key[i]), value[i])
	}
	return nil
}

// PutIfNotExists inserts a <key, value> record only if it does not exist yet, otherwise return ErrAlreadyExist
func (m *memKVStore) PutIfNotExists(namespace string, key []byte, value []byte) error {
	_, ok := m.data.Load(namespace + keyDelimiter + string(key))
	if !ok {
		m.data.Store(namespace+keyDelimiter+string(key), value)
		return nil
	}
	return ErrAlreadyExist
}

// Get retrieves a record
func (m *memKVStore) Get(namespace string, key []byte) ([]byte, error) {
	value, _ := m.data.Load(namespace + keyDelimiter + string(key))
	if value != nil {
		return value.([]byte), nil
	}
	return nil, errors.Wrapf(ErrNotExist, "key = %x", key)
}

// Delete deletes a record
func (m *memKVStore) Delete(namespace string, key []byte) error {
	m.data.Delete(namespace + keyDelimiter + string(key))
	return nil
}

const (
	fileMode = 0600
)

// boltDB is KVStore implementation based bolt DB
type boltDB struct {
	service.AbstractService
	db      *bolt.DB
	path    string
	options *bolt.Options
}

// NewBoltDB instantiates a boltdb based KV store
func NewBoltDB(path string, options *bolt.Options) KVStore {
	return &boltDB{path: path, options: options}
}

// Start opens the BoltDB (creates new file if not existing yet)
func (b *boltDB) Start() error {
	db, err := bolt.Open(b.path, fileMode, b.options)
	if err != nil {
		return err
	}
	b.db = db
	return nil
}

// Stop closes the BoltDB
func (b *boltDB) Stop() error {
	return b.db.Close()
}

// Put inserts a <key, value> record
func (b *boltDB) Put(namespace string, key []byte, value []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(namespace))
		if err != nil {
			return err
		}
		return bucket.Put(key, value)
	})
}

// BatchPut inserts a slice of records <key[], value[]>
func (b *boltDB) BatchPut(namespace string, key [][]byte, value [][]byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(namespace))
		if err != nil {
			return err
		}
		if len(key) != len(value) {
			return errors.Wrap(ErrInvalidDB, "batch put <k, v> size not match")
		}
		for i := 0; i < len(key); i++ {
			if err := bucket.Put(key[i], value[i]); err != nil {
				return err
			}
		}
		return nil
	})
}

// PutIfNotExists inserts a <key, value> record only if it does not exist yet, otherwise return ErrAlreadyExist
func (b *boltDB) PutIfNotExists(namespace string, key []byte, value []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(namespace))
		if err != nil {
			return err
		}
		if bucket.Get(key) == nil {
			return bucket.Put(key, value)
		}
		return ErrAlreadyExist
	})
}

// Get retrieves a record
func (b *boltDB) Get(namespace string, key []byte) ([]byte, error) {
	var value []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(namespace))
		if bucket == nil {
			return errors.Wrapf(bolt.ErrBucketNotFound, "bucket = %s", namespace)
		}
		value = bucket.Get(key)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if value == nil {
		err = errors.Wrapf(ErrNotExist, "key = %x", key)
	}
	return value, err
}

// Delete deletes a record
func (b *boltDB) Delete(namespace string, key []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(namespace))
		if bucket == nil {
			return errors.Wrapf(bolt.ErrBucketNotFound, "bucket = %s", namespace)
		}
		return bucket.Delete(key)
	})
}

//======================================
// private functions
//======================================

// intentionally fail to test DB can successfully rollback
func (b *boltDB) batchPutForceFail(namespace string, key [][]byte, value [][]byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(namespace))
		if err != nil {
			return err
		}
		if len(key) != len(value) {
			return errors.Wrap(ErrInvalidDB, "batch put <k, v> size not match")
		}
		for i := 0; i < len(key); i++ {
			if err := bucket.Put(key[i], value[i]); err != nil {
				return err
			}
			// intentionally fail to test DB can successfully rollback
			if i == len(key)-1 {
				return errors.Wrapf(ErrInvalidDB, "force fail to test DB rollback")
			}
		}
		return nil
	})
}
