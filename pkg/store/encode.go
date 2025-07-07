package store

import (
	"encoding/binary"
	"fmt"
)

// EncodeConfiguration encodes a Configuration into the format expected by HashiCorp Raft
func EncodeConfiguration(config Configuration) ([]byte, error) {
	// Calculate size
	size := 1 // protocol version
	for _, server := range config.Servers {
		size += 1                   // suffrage
		size += 8                   // id length
		size += len(server.ID)      // id
		size += 8                   // address length
		size += len(server.Address) // address
	}

	buf := make([]byte, 0, size)

	// Protocol version (1)
	buf = append(buf, 1)

	// Encode servers
	for _, server := range config.Servers {
		// Suffrage
		buf = append(buf, server.Suffrage)

		// ID length and value
		idLen := uint64(len(server.ID))
		buf = binary.BigEndian.AppendUint64(buf, idLen)
		buf = append(buf, []byte(server.ID)...)

		// Address length and value
		addrLen := uint64(len(server.Address))
		buf = binary.BigEndian.AppendUint64(buf, addrLen)
		buf = append(buf, []byte(server.Address)...)
	}

	return buf, nil
}

// encodeLogEntry encodes a LogEntry into the format expected by HashiCorp Raft
func encodeLogEntry(entry *LogEntry) ([]byte, error) {
	// Calculate total size
	size := 8 + 8 + 1 + len(entry.Data) // term + index + type + data
	buf := make([]byte, 0, size)

	// Term (8 bytes)
	buf = binary.BigEndian.AppendUint64(buf, entry.Term)

	// Index (8 bytes)
	buf = binary.BigEndian.AppendUint64(buf, entry.Index)

	// Type (1 byte)
	buf = append(buf, entry.Type)

	// Data
	buf = append(buf, entry.Data...)

	return buf, nil
}

// decodeConfiguration decodes a Configuration from bytes
func decodeConfiguration(data []byte) (*Configuration, error) {
	if len(data) < 1 {
		return nil, fmt.Errorf("configuration data too short")
	}

	// Check protocol version
	version := data[0]
	if version != 1 {
		return nil, fmt.Errorf("unsupported configuration version: %d", version)
	}

	config := &Configuration{
		Servers: make([]Server, 0),
	}

	idx := 1
	for idx < len(data) {
		if idx+1 > len(data) {
			return nil, fmt.Errorf("incomplete server entry")
		}

		server := Server{
			Suffrage: data[idx],
		}
		idx++

		// Read ID
		if idx+8 > len(data) {
			return nil, fmt.Errorf("incomplete ID length")
		}
		idLen := binary.BigEndian.Uint64(data[idx : idx+8])
		idx += 8

		if idx+int(idLen) > len(data) {
			return nil, fmt.Errorf("incomplete ID data")
		}
		server.ID = string(data[idx : idx+int(idLen)])
		idx += int(idLen)

		// Read Address
		if idx+8 > len(data) {
			return nil, fmt.Errorf("incomplete address length")
		}
		addrLen := binary.BigEndian.Uint64(data[idx : idx+8])
		idx += 8

		if idx+int(addrLen) > len(data) {
			return nil, fmt.Errorf("incomplete address data")
		}
		server.Address = string(data[idx : idx+int(addrLen)])
		idx += int(addrLen)

		config.Servers = append(config.Servers, server)
	}

	return config, nil
}
