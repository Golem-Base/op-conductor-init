package store

import (
	"fmt"
	"path/filepath"

	bolt "go.etcd.io/bbolt"
)

const (
	// Bucket names used by HashiCorp Raft
	dbLogs = "logs"
	dbConf = "conf"
)

// BoltStore wraps a BoltDB database for Raft storage
type BoltStore struct {
	db   *bolt.DB
	path string
}

// NewBoltStore creates a new BoltStore
func NewBoltStore(path string) (*BoltStore, error) {
	db, err := bolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open bolt db: %w", err)
	}

	store := &BoltStore{
		db:   db,
		path: path,
	}

	// Initialize buckets
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(dbLogs)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(dbConf)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create buckets: %w", err)
	}

	return store, nil
}

// Close closes the BoltDB database
func (s *BoltStore) Close() error {
	return s.db.Close()
}

// StoreLog stores a log entry
func (s *BoltStore) StoreLog(entry *LogEntry) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(dbLogs))
		if bucket == nil {
			return fmt.Errorf("logs bucket not found")
		}

		key := uint64ToBytes(entry.Index)
		value, err := encodeLogEntry(entry)
		if err != nil {
			return fmt.Errorf("failed to encode log entry: %w", err)
		}

		return bucket.Put(key, value)
	})
}

// SetUint64 sets a uint64 value in the conf bucket
func (s *BoltStore) SetUint64(key string, value uint64) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(dbConf))
		if bucket == nil {
			return fmt.Errorf("conf bucket not found")
		}
		return bucket.Put([]byte(key), uint64ToBytes(value))
	})
}

// Set sets a value in the conf bucket
func (s *BoltStore) Set(key, value []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(dbConf))
		if bucket == nil {
			return fmt.Errorf("conf bucket not found")
		}
		return bucket.Put(key, value)
	})
}

// CreateStableStore creates the stable store for a node
func CreateStableStore(nodeDir string, serverID string, initialTerm uint64, isLeader bool) error {
	dbPath := filepath.Join(nodeDir, "raft-stable.db")
	store, err := NewBoltStore(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create stable store: %w", err)
	}
	defer store.Close()

	// Set current term
	if err := store.SetUint64("CurrentTerm", initialTerm); err != nil {
		return fmt.Errorf("failed to set current term: %w", err)
	}

	// Set last vote (only for leader)
	if isLeader {
		if err := store.Set([]byte("LastVoteTerm"), uint64ToBytes(initialTerm)); err != nil {
			return fmt.Errorf("failed to set last vote term: %w", err)
		}
		if err := store.Set([]byte("LastVoteCand"), []byte(serverID)); err != nil {
			return fmt.Errorf("failed to set last vote candidate: %w", err)
		}
	}

	return nil
}

// CreateLogStore creates the log store for a node with initial configuration
func CreateLogStore(nodeDir string, configEntry *LogEntry) error {
	dbPath := filepath.Join(nodeDir, "raft-log.db")
	store, err := NewBoltStore(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create log store: %w", err)
	}
	defer store.Close()

	// Store the configuration entry
	if err := store.StoreLog(configEntry); err != nil {
		return fmt.Errorf("failed to store configuration entry: %w", err)
	}

	// Store FirstIndex and LastIndex for fast lookup
	if err := store.SetUint64("FirstIndex", configEntry.Index); err != nil {
		return fmt.Errorf("failed to set first index: %w", err)
	}
	if err := store.SetUint64("LastIndex", configEntry.Index); err != nil {
		return fmt.Errorf("failed to set last index: %w", err)
	}

	return nil
}
