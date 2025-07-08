package store

import (
	"encoding/binary"
)

// Constants for Raft log types (must match github.com/hashicorp/raft)
const (
	LogCommand              uint8 = 0
	LogNoop                 uint8 = 1
	LogAddPeerDeprecated    uint8 = 2
	LogRemovePeerDeprecated uint8 = 3
	LogBarrier              uint8 = 4
	LogConfiguration        uint8 = 5
)

// Constants for server suffrage
const (
	Voter    uint8 = 0
	Nonvoter uint8 = 1
	Staging  uint8 = 2
)

// LogEntry represents a Raft log entry
type LogEntry struct {
	Index uint64
	Term  uint64
	Type  uint8
	Data  []byte
}

// Server represents a server in the Raft configuration
type Server struct {
	Suffrage uint8
	ID       string
	Address  string
}

// Configuration represents a Raft configuration
type Configuration struct {
	Servers []Server
}

// ConfigurationEntry represents a configuration change log entry
type ConfigurationEntry struct {
	Configuration Configuration
}

// uint64ToBytes converts a uint64 to big-endian byte representation
func uint64ToBytes(n uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, n)
	return buf
}

// bytesToUint64 converts big-endian bytes to uint64
func bytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}
