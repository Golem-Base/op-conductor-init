package store

import (
	"bytes"
	"testing"

	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/raft"
)

// TestLogEntryEncodingCompatibility tests that our encoding matches raft-boltdb
func TestLogEntryEncodingCompatibility(t *testing.T) {
	// Create a test log entry
	logEntry := &LogEntry{
		Index: 1,
		Term:  1,
		Type:  LogConfiguration,
		Data:  []byte("test configuration data"),
	}

	// Encode using our method
	encoded, err := encodeLogEntry(logEntry)
	if err != nil {
		t.Fatalf("Failed to encode log entry: %v", err)
	}

	// Decode using raft's msgpack (simulating what raft-boltdb does)
	var decodedLog raft.Log
	handle := codec.MsgpackHandle{}
	decoder := codec.NewDecoder(bytes.NewReader(encoded), &handle)
	if err := decoder.Decode(&decodedLog); err != nil {
		t.Fatalf("Failed to decode log entry: %v", err)
	}

	// Verify fields match
	if decodedLog.Index != logEntry.Index {
		t.Errorf("Index mismatch: got %d, want %d", decodedLog.Index, logEntry.Index)
	}
	if decodedLog.Term != logEntry.Term {
		t.Errorf("Term mismatch: got %d, want %d", decodedLog.Term, logEntry.Term)
	}
	if decodedLog.Type != raft.LogType(logEntry.Type) {
		t.Errorf("Type mismatch: got %d, want %d", decodedLog.Type, logEntry.Type)
	}
	if !bytes.Equal(decodedLog.Data, logEntry.Data) {
		t.Errorf("Data mismatch: got %v, want %v", decodedLog.Data, logEntry.Data)
	}
}

// TestConfigurationEncodingRoundtrip tests configuration encoding/decoding
func TestConfigurationEncodingRoundtrip(t *testing.T) {
	// Create test configuration
	config := Configuration{
		Servers: []Server{
			{ID: "server1", Address: "localhost:8080", Suffrage: Voter},
			{ID: "server2", Address: "localhost:8081", Suffrage: Voter},
			{ID: "server3", Address: "localhost:8082", Suffrage: Nonvoter},
		},
	}

	// Encode
	encoded, err := EncodeConfiguration(config)
	if err != nil {
		t.Fatalf("Failed to encode configuration: %v", err)
	}

	// Decode
	decoded, err := decodeConfiguration(encoded)
	if err != nil {
		t.Fatalf("Failed to decode configuration: %v", err)
	}

	// Verify
	if len(decoded.Servers) != len(config.Servers) {
		t.Fatalf("Server count mismatch: got %d, want %d", len(decoded.Servers), len(config.Servers))
	}

	for i, server := range config.Servers {
		if decoded.Servers[i].ID != server.ID {
			t.Errorf("Server %d ID mismatch: got %s, want %s", i, decoded.Servers[i].ID, server.ID)
		}
		if decoded.Servers[i].Address != server.Address {
			t.Errorf("Server %d Address mismatch: got %s, want %s", i, decoded.Servers[i].Address, server.Address)
		}
		if decoded.Servers[i].Suffrage != server.Suffrage {
			t.Errorf("Server %d Suffrage mismatch: got %d, want %d", i, decoded.Servers[i].Suffrage, server.Suffrage)
		}
	}
}

// TestRaftConfigurationCompatibility tests that our configuration is compatible with raft.Configuration
func TestRaftConfigurationCompatibility(t *testing.T) {
	config := Configuration{
		Servers: []Server{
			{ID: "node1", Address: "10.0.0.1:50050", Suffrage: Voter},
			{ID: "node2", Address: "10.0.0.2:50050", Suffrage: Voter},
		},
	}

	// Encode our configuration
	encoded, err := EncodeConfiguration(config)
	if err != nil {
		t.Fatalf("Failed to encode configuration: %v", err)
	}

	// Decode as raft.Configuration (what op-conductor will do)
	raftConfig := raft.DecodeConfiguration(encoded)

	// Verify the configuration
	if len(raftConfig.Servers) != len(config.Servers) {
		t.Fatalf("Server count mismatch: got %d, want %d", len(raftConfig.Servers), len(config.Servers))
	}

	for i, server := range config.Servers {
		raftServer := raftConfig.Servers[i]
		if string(raftServer.ID) != server.ID {
			t.Errorf("Server %d ID mismatch: got %s, want %s", i, raftServer.ID, server.ID)
		}
		if string(raftServer.Address) != server.Address {
			t.Errorf("Server %d Address mismatch: got %s, want %s", i, raftServer.Address, server.Address)
		}
		if uint8(raftServer.Suffrage) != server.Suffrage {
			t.Errorf("Server %d Suffrage mismatch: got %d, want %d", i, raftServer.Suffrage, server.Suffrage)
		}
	}
}

// TestLogEntryWithConfiguration tests encoding a configuration log entry
func TestLogEntryWithConfiguration(t *testing.T) {
	// Create configuration
	config := Configuration{
		Servers: []Server{
			{ID: "sequencer-1-0", Address: "sequencer-1:50050", Suffrage: Voter},
			{ID: "sequencer-2-0", Address: "sequencer-2:50050", Suffrage: Voter},
			{ID: "sequencer-3-0", Address: "sequencer-3:50050", Suffrage: Voter},
		},
	}

	// Encode configuration
	configData, err := EncodeConfiguration(config)
	if err != nil {
		t.Fatalf("Failed to encode configuration: %v", err)
	}

	// Create log entry with configuration
	logEntry := &LogEntry{
		Index: 1,
		Term:  1,
		Type:  LogConfiguration,
		Data:  configData,
	}

	// Encode log entry
	encoded, err := encodeLogEntry(logEntry)
	if err != nil {
		t.Fatalf("Failed to encode log entry: %v", err)
	}

	// Decode as raft.Log
	var decodedLog raft.Log
	handle := codec.MsgpackHandle{}
	decoder := codec.NewDecoder(bytes.NewReader(encoded), &handle)
	if err := decoder.Decode(&decodedLog); err != nil {
		t.Fatalf("Failed to decode log entry: %v", err)
	}

	// Verify log metadata
	if decodedLog.Index != 1 {
		t.Errorf("Index mismatch: got %d, want 1", decodedLog.Index)
	}
	if decodedLog.Term != 1 {
		t.Errorf("Term mismatch: got %d, want 1", decodedLog.Term)
	}
	if decodedLog.Type != raft.LogConfiguration {
		t.Errorf("Type mismatch: got %d, want %d (LogConfiguration)", decodedLog.Type, raft.LogConfiguration)
	}

	// Decode the configuration from the log data
	decodedConfig := raft.DecodeConfiguration(decodedLog.Data)
	if len(decodedConfig.Servers) != 3 {
		t.Fatalf("Expected 3 servers, got %d", len(decodedConfig.Servers))
	}

	// Verify it contains our servers
	expectedServers := map[string]string{
		"sequencer-1-0": "sequencer-1:50050",
		"sequencer-2-0": "sequencer-2:50050",
		"sequencer-3-0": "sequencer-3:50050",
	}

	for _, server := range decodedConfig.Servers {
		expectedAddr, ok := expectedServers[string(server.ID)]
		if !ok {
			t.Errorf("Unexpected server ID: %s", server.ID)
		}
		if string(server.Address) != expectedAddr {
			t.Errorf("Address mismatch for %s: got %s, want %s", server.ID, server.Address, expectedAddr)
		}
		if server.Suffrage != raft.Voter {
			t.Errorf("Expected voter suffrage for %s, got %v", server.ID, server.Suffrage)
		}
	}
}
