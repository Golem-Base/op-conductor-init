package store

import (
	"os"
	"path/filepath"
	"testing"

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
			bucket := tx.Bucket([]byte(dbConf))
			if bucket == nil {
				t.Fatal("conf bucket not found")
			}

			// Check CurrentTerm
			termBytes := bucket.Get([]byte("CurrentTerm"))
			if termBytes == nil {
				t.Fatal("CurrentTerm not found")
			}
			term := bytesToUint64(termBytes)
			if term != 1 {
				t.Fatalf("Expected term 1, got %d", term)
			}

			// Check LastVoteCand (should exist for leader)
			vote := bucket.Get([]byte("LastVoteCand"))
			if string(vote) != "test-server" {
				t.Fatalf("Expected vote for test-server, got %s", vote)
			}

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	// Test log store
	t.Run("LogStore", func(t *testing.T) {
		configEntry := &LogEntry{
			Index: 1,
			Term:  1,
			Type:  LogConfiguration,
			Data:  []byte("test-config"),
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
			bucket := tx.Bucket([]byte(dbLogs))
			if bucket == nil {
				t.Fatal("logs bucket not found")
			}

			// Check log entry
			entryBytes := bucket.Get(uint64ToBytes(1))
			if entryBytes == nil {
				t.Fatal("Log entry not found")
			}

			// Basic verification - should have term, index, type, and data
			if len(entryBytes) < 17 { // 8 (term) + 8 (index) + 1 (type)
				t.Fatal("Log entry too short")
			}

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestEncodeDecodeConfiguration(t *testing.T) {
	config := Configuration{
		Servers: []Server{
			{Suffrage: Voter, ID: "server-1", Address: "addr-1:50050"},
			{Suffrage: Voter, ID: "server-2", Address: "addr-2:50050"},
			{Suffrage: Voter, ID: "server-3", Address: "addr-3:50050"},
		},
	}

	// Encode
	data, err := EncodeConfiguration(config)
	if err != nil {
		t.Fatalf("Failed to encode configuration: %v", err)
	}

	// Decode
	decoded, err := decodeConfiguration(data)
	if err != nil {
		t.Fatalf("Failed to decode configuration: %v", err)
	}

	// Verify
	if len(decoded.Servers) != len(config.Servers) {
		t.Fatalf("Server count mismatch: expected %d, got %d",
			len(config.Servers), len(decoded.Servers))
	}

	for i, server := range decoded.Servers {
		expected := config.Servers[i]
		if server.ID != expected.ID {
			t.Errorf("Server %d ID mismatch: expected %s, got %s",
				i, expected.ID, server.ID)
		}
		if server.Address != expected.Address {
			t.Errorf("Server %d Address mismatch: expected %s, got %s",
				i, expected.Address, server.Address)
		}
		if server.Suffrage != expected.Suffrage {
			t.Errorf("Server %d Suffrage mismatch: expected %d, got %d",
				i, expected.Suffrage, server.Suffrage)
		}
	}
}
