package store

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/raft"
	bolt "go.etcd.io/bbolt"
)

func TestBoltStore(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "raft-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test creating stores
	nodeDir := filepath.Join(tmpDir, "test-node")
	os.MkdirAll(nodeDir, 0o755)

	// Test stable store
	t.Run("StableStore", func(t *testing.T) {
		err := CreateStableStore(nodeDir, "test-server", 1, true)
		if err != nil {
			t.Fatalf("Failed to create stable store: %v", err)
		}

		// Verify the contents
		db, err := bolt.Open(filepath.Join(nodeDir, "raft-stable.db"), 0o600, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		err = db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("conf"))
			if bucket == nil {
				t.Fatal("conf bucket not found")
			}

			// Check CurrentTerm
			termBytes := bucket.Get([]byte("CurrentTerm"))
			if termBytes == nil {
				t.Fatal("CurrentTerm not found")
			}
			term := binary.BigEndian.Uint64(termBytes)
			if term != 1 {
				t.Fatalf("Expected term 1, got %d", term)
			}

			// Check LastVoteTerm (should exist for leader)
			voteTermBytes := bucket.Get([]byte("LastVoteTerm"))
			if voteTermBytes == nil {
				t.Fatal("LastVoteTerm not found for leader")
			}

			// Check LastVoteCand (should exist for leader)
			voteCand := bucket.Get([]byte("LastVoteCand"))
			if voteCand == nil {
				t.Fatal("LastVoteCand not found for leader")
			}
			if string(voteCand) != "test-server" {
				t.Fatalf("Expected LastVoteCand to be test-server, got %s", string(voteCand))
			}

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	// Test log store
	t.Run("LogStore", func(t *testing.T) {
		// Create a configuration entry
		config := raft.Configuration{
			Servers: []raft.Server{
				{
					Suffrage: raft.Voter,
					ID:       "server1",
					Address:  "127.0.0.1:8300",
				},
				{
					Suffrage: raft.Voter,
					ID:       "server2",
					Address:  "127.0.0.1:8301",
				},
			},
		}

		configEntry := &raft.Log{
			Index: 1,
			Term:  1,
			Type:  raft.LogConfiguration,
			Data:  raft.EncodeConfiguration(config),
		}

		err := CreateLogStore(nodeDir, configEntry)
		if err != nil {
			t.Fatalf("Failed to create log store: %v", err)
		}

		// Verify the contents
		db, err := bolt.Open(filepath.Join(nodeDir, "raft-log.db"), 0o600, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		err = db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("logs"))
			if bucket == nil {
				t.Fatal("logs bucket not found")
			}

			// Check that entry exists
			key := make([]byte, 8)
			binary.BigEndian.PutUint64(key, 1)
			entry := bucket.Get(key)
			if entry == nil {
				t.Fatal("Log entry not found")
			}

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestStableStoreFollower(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "raft-follower-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	nodeDir := filepath.Join(tmpDir, "follower-node")
	os.MkdirAll(nodeDir, 0o755)

	// Create stable store for follower (isLeader = false)
	err = CreateStableStore(nodeDir, "follower-server", 1, false)
	if err != nil {
		t.Fatalf("Failed to create stable store: %v", err)
	}

	// Verify follower doesn't have vote information
	db, err := bolt.Open(filepath.Join(nodeDir, "raft-stable.db"), 0o600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("conf"))
		if bucket == nil {
			t.Fatal("conf bucket not found")
		}

		// Check CurrentTerm exists
		termBytes := bucket.Get([]byte("CurrentTerm"))
		if termBytes == nil {
			t.Fatal("CurrentTerm not found")
		}

		// Check LastVoteTerm doesn't exist for follower
		voteTermBytes := bucket.Get([]byte("LastVoteTerm"))
		if voteTermBytes != nil {
			t.Fatal("LastVoteTerm should not exist for follower")
		}

		// Check LastVoteCand doesn't exist for follower
		voteCand := bucket.Get([]byte("LastVoteCand"))
		if voteCand != nil {
			t.Fatal("LastVoteCand should not exist for follower")
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
