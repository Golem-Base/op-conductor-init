package store

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb/v2"
)

// CreateStableStore creates the stable store for a node using HashiCorp's raft-boltdb
func CreateStableStore(nodeDir string, serverID string, initialTerm uint64, isLeader bool) error {
	dbPath := filepath.Join(nodeDir, "raft-stable.db")

	// Create stable store using HashiCorp's raft-boltdb
	store, err := boltdb.NewBoltStore(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create stable store: %w", err)
	}
	defer store.Close()

	// Set current term
	if err := store.SetUint64([]byte("CurrentTerm"), initialTerm); err != nil {
		return fmt.Errorf("failed to set current term: %w", err)
	}

	// Set last vote (only for leader)
	if isLeader {
		if err := store.SetUint64([]byte("LastVoteTerm"), initialTerm); err != nil {
			return fmt.Errorf("failed to set last vote term: %w", err)
		}
		if err := store.Set([]byte("LastVoteCand"), []byte(serverID)); err != nil {
			return fmt.Errorf("failed to set last vote candidate: %w", err)
		}
	}

	return nil
}

// CreateLogStore creates the log store for a node with initial configuration
func CreateLogStore(nodeDir string, configEntry *raft.Log) error {
	dbPath := filepath.Join(nodeDir, "raft-log.db")

	// Create log store using HashiCorp's raft-boltdb
	store, err := boltdb.NewBoltStore(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create log store: %w", err)
	}
	defer store.Close()

	// Store the configuration entry
	if err := store.StoreLog(configEntry); err != nil {
		return fmt.Errorf("failed to store configuration entry: %w", err)
	}

	// The raft-boltdb library automatically maintains FirstIndex and LastIndex
	// No need to set them manually

	return nil
}
