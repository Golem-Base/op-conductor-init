package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb/v2"
)

// TestRaftBoltDBCompatibility verifies that our bolt database format is compatible
// with the official hashicorp/raft-boltdb library, ensuring op-conductor can read
// the files we generate
func TestRaftBoltDBCompatibility(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "raft-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create our bolt store and populate it
	dbPath := filepath.Join(tmpDir, "raft-log.db")
	ourStore, err := NewBoltStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create our store: %v", err)
	}

	// Create test configuration
	config := Configuration{
		Servers: []Server{
			{ID: "node1", Address: "10.0.0.1:50050", Suffrage: Voter},
			{ID: "node2", Address: "10.0.0.2:50050", Suffrage: Voter},
			{ID: "node3", Address: "10.0.0.3:50050", Suffrage: Nonvoter},
		},
	}

	configData, err := EncodeConfiguration(config)
	if err != nil {
		t.Fatalf("Failed to encode configuration: %v", err)
	}

	// Store configuration log entry
	logEntry := &LogEntry{
		Index: 1,
		Term:  1,
		Type:  LogConfiguration,
		Data:  configData,
	}

	if err := ourStore.StoreLog(logEntry); err != nil {
		t.Fatalf("Failed to store log: %v", err)
	}

	// Set indices
	if err := ourStore.SetUint64("FirstIndex", 1); err != nil {
		t.Fatalf("Failed to set FirstIndex: %v", err)
	}
	if err := ourStore.SetUint64("LastIndex", 1); err != nil {
		t.Fatalf("Failed to set LastIndex: %v", err)
	}

	// Close our store
	if err := ourStore.Close(); err != nil {
		t.Fatalf("Failed to close our store: %v", err)
	}

	// Now open with raft-boltdb and verify
	raftStore, err := boltdb.NewBoltStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to open with raft-boltdb: %v", err)
	}
	defer raftStore.Close()

	// Check indices
	firstIdx, err := raftStore.FirstIndex()
	if err != nil {
		t.Fatalf("Failed to get first index: %v", err)
	}
	if firstIdx != 1 {
		t.Errorf("FirstIndex mismatch: got %d, want 1", firstIdx)
	}

	lastIdx, err := raftStore.LastIndex()
	if err != nil {
		t.Fatalf("Failed to get last index: %v", err)
	}
	if lastIdx != 1 {
		t.Errorf("LastIndex mismatch: got %d, want 1", lastIdx)
	}

	// Get the log entry
	var log raft.Log
	if err := raftStore.GetLog(1, &log); err != nil {
		t.Fatalf("Failed to get log: %v", err)
	}

	// Verify log metadata
	if log.Index != 1 {
		t.Errorf("Log index mismatch: got %d, want 1", log.Index)
	}
	if log.Term != 1 {
		t.Errorf("Log term mismatch: got %d, want 1", log.Term)
	}
	if log.Type != raft.LogConfiguration {
		t.Errorf("Log type mismatch: got %d, want %d", log.Type, raft.LogConfiguration)
	}

	// Decode and verify configuration
	decodedConfig := raft.DecodeConfiguration(log.Data)
	if len(decodedConfig.Servers) != 3 {
		t.Fatalf("Server count mismatch: got %d, want 3", len(decodedConfig.Servers))
	}

	// Verify each server
	expectedServers := []struct {
		ID       string
		Address  string
		Suffrage raft.ServerSuffrage
	}{
		{"node1", "10.0.0.1:50050", raft.Voter},
		{"node2", "10.0.0.2:50050", raft.Voter},
		{"node3", "10.0.0.3:50050", raft.Nonvoter},
	}

	for i, expected := range expectedServers {
		server := decodedConfig.Servers[i]
		if string(server.ID) != expected.ID {
			t.Errorf("Server %d ID mismatch: got %s, want %s", i, server.ID, expected.ID)
		}
		if string(server.Address) != expected.Address {
			t.Errorf("Server %d Address mismatch: got %s, want %s", i, server.Address, expected.Address)
		}
		if server.Suffrage != expected.Suffrage {
			t.Errorf("Server %d Suffrage mismatch: got %v, want %v", i, server.Suffrage, expected.Suffrage)
		}
	}
}

// TestStableStoreCompatibility verifies that our stable store format matches
// what raft-boltdb expects for storing term and voting information
func TestStableStoreCompatibility(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "raft-stable-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create stable store using our code
	dbPath := filepath.Join(tmpDir, "raft-stable.db")
	ourStore, err := NewBoltStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create our store: %v", err)
	}

	// Set values
	if err := ourStore.Set([]byte("CurrentTerm"), uint64ToBytes(5)); err != nil {
		t.Fatalf("Failed to set CurrentTerm: %v", err)
	}
	if err := ourStore.Set([]byte("LastVoteTerm"), uint64ToBytes(5)); err != nil {
		t.Fatalf("Failed to set LastVoteTerm: %v", err)
	}
	if err := ourStore.Set([]byte("LastVoteCand"), []byte("node1")); err != nil {
		t.Fatalf("Failed to set LastVoteCand: %v", err)
	}

	if err := ourStore.Close(); err != nil {
		t.Fatalf("Failed to close our store: %v", err)
	}

	// Open with raft-boltdb and verify
	raftStore, err := boltdb.NewBoltStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to open with raft-boltdb: %v", err)
	}
	defer raftStore.Close()

	// Get CurrentTerm
	term, err := raftStore.GetUint64([]byte("CurrentTerm"))
	if err != nil {
		t.Fatalf("Failed to get CurrentTerm: %v", err)
	}
	if term != 5 {
		t.Errorf("CurrentTerm mismatch: got %d, want 5", term)
	}

	// Get LastVoteTerm
	voteTerm, err := raftStore.GetUint64([]byte("LastVoteTerm"))
	if err != nil {
		t.Fatalf("Failed to get LastVoteTerm: %v", err)
	}
	if voteTerm != 5 {
		t.Errorf("LastVoteTerm mismatch: got %d, want 5", voteTerm)
	}

	// Get LastVoteCand
	voteCand, err := raftStore.Get([]byte("LastVoteCand"))
	if err != nil {
		t.Fatalf("Failed to get LastVoteCand: %v", err)
	}
	if string(voteCand) != "node1" {
		t.Errorf("LastVoteCand mismatch: got %s, want node1", string(voteCand))
	}
}
