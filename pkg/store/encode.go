package store

import (
	"fmt"

	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/raft"
)

// EncodeConfiguration encodes a Configuration into the format expected by HashiCorp Raft
func EncodeConfiguration(config Configuration) ([]byte, error) {
	// Convert to HashiCorp Raft Configuration
	raftConfig := raft.Configuration{
		Servers: make([]raft.Server, len(config.Servers)),
	}

	for i, server := range config.Servers {
		raftConfig.Servers[i] = raft.Server{
			ID:       raft.ServerID(server.ID),
			Address:  raft.ServerAddress(server.Address),
			Suffrage: raft.ServerSuffrage(server.Suffrage),
		}
	}

	// Use Raft's built-in encoding
	return raft.EncodeConfiguration(raftConfig), nil
}

// decodeConfiguration decodes a Configuration from bytes
func decodeConfiguration(data []byte) (*Configuration, error) {
	raftConfig := raft.DecodeConfiguration(data)

	config := &Configuration{
		Servers: make([]Server, len(raftConfig.Servers)),
	}

	for i, server := range raftConfig.Servers {
		config.Servers[i] = Server{
			ID:       string(server.ID),
			Address:  string(server.Address),
			Suffrage: uint8(server.Suffrage),
		}
	}

	return config, nil
}

// encodeLogEntry encodes a log entry using msgpack format expected by HashiCorp Raft
func encodeLogEntry(entry *LogEntry) ([]byte, error) {
	raftLog := &raft.Log{
		Index:      entry.Index,
		Term:       entry.Term,
		Type:       raft.LogType(entry.Type),
		Data:       entry.Data,
		Extensions: nil,
		// AppendedAt is automatically set to zero value
	}

	var handle codec.MsgpackHandle

	var buf []byte
	encoder := codec.NewEncoderBytes(&buf, &handle)
	if err := encoder.Encode(raftLog); err != nil {
		return nil, fmt.Errorf("failed to msgpack encode log entry: %w", err)
	}

	return buf, nil
}
