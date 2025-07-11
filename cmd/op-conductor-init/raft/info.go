package raft

import (
	"encoding/binary"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/raft"
	"github.com/urfave/cli/v2"
	bolt "go.etcd.io/bbolt"
)

// InfoAction handles the info subcommand
func InfoAction(ctx *cli.Context) error {
	stateDir := ctx.String("state-dir")

	fmt.Printf("=== Raft State Information ===\n")
	fmt.Printf("Directory: %s\n\n", stateDir)

	// Check stable store
	fmt.Println("Stable Store (raft-stable.db):")
	fmt.Println("------------------------------")
	stablePath := filepath.Join(stateDir, "raft-stable.db")
	if err := showStableStoreInfo(stablePath); err != nil {
		return fmt.Errorf("error reading stable store: %w", err)
	}

	fmt.Println("\nLog Store (raft-log.db):")
	fmt.Println("------------------------")
	logPath := filepath.Join(stateDir, "raft-log.db")
	if err := showLogStoreInfo(logPath); err != nil {
		return fmt.Errorf("error reading log store: %w", err)
	}

	return nil
}

func showStableStoreInfo(path string) error {
	db, err := bolt.Open(path, 0o600, &bolt.Options{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("failed to open stable store: %w", err)
	}
	defer db.Close()

	return db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("conf"))
		if bucket == nil {
			return fmt.Errorf("conf bucket not found")
		}

		var currentTerm, lastVoteTerm uint64
		var lastVoteCand string
		var hasVoted bool

		// Read all values
		bucket.ForEach(func(k, v []byte) error {
			key := string(k)
			switch key {
			case "CurrentTerm":
				if len(v) == 8 {
					currentTerm = binary.BigEndian.Uint64(v)
				}
			case "LastVoteTerm":
				if len(v) == 8 {
					lastVoteTerm = binary.BigEndian.Uint64(v)
					hasVoted = true
				}
			case "LastVoteCand":
				lastVoteCand = string(v)
			}
			return nil
		})

		fmt.Printf("  Current Term: %d\n", currentTerm)
		if hasVoted {
			fmt.Printf("  Last Vote Term: %d\n", lastVoteTerm)
			fmt.Printf("  Last Vote Candidate: %s\n", lastVoteCand)
			fmt.Printf("  Node Role: %s\n", determineRole(lastVoteCand))
		} else {
			fmt.Printf("  Node Role: Follower (no vote recorded)\n")
		}

		return nil
	})
}

func showLogStoreInfo(path string) error {
	db, err := bolt.Open(path, 0o600, &bolt.Options{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("failed to open log store: %w", err)
	}
	defer db.Close()

	return db.View(func(tx *bolt.Tx) error {
		// Check conf bucket
		confBucket := tx.Bucket([]byte("conf"))
		if confBucket != nil {
			var firstIndex, lastIndex uint64
			if v := confBucket.Get([]byte("FirstIndex")); len(v) == 8 {
				firstIndex = binary.BigEndian.Uint64(v)
			}
			if v := confBucket.Get([]byte("LastIndex")); len(v) == 8 {
				lastIndex = binary.BigEndian.Uint64(v)
			}
			fmt.Printf("  First Index: %d\n", firstIndex)
			fmt.Printf("  Last Index: %d\n", lastIndex)
			fmt.Printf("  Total Entries: %d\n", lastIndex-firstIndex+1)
		}

		// Check logs bucket
		logsBucket := tx.Bucket([]byte("logs"))
		if logsBucket == nil {
			return fmt.Errorf("logs bucket not found")
		}

		fmt.Println("\n  Log Entries:")
		entryCount := 0
		var clusterMembers []string

		logsBucket.ForEach(func(k, v []byte) error {
			if len(k) == 8 && len(v) >= 17 {
				index := binary.BigEndian.Uint64(k)
				term := binary.BigEndian.Uint64(v[0:8])
				// logIndex := binary.BigEndian.Uint64(v[8:16])  // Not used currently
				logType := v[16]
				dataLen := len(v) - 17

				fmt.Printf("    Entry %d: term=%d, type=%s, size=%d bytes\n",
					index, term, getLogTypeName(logType), dataLen)

				// If it's a configuration entry, decode it
				if logType == uint8(raft.LogConfiguration) && len(v) > 17 {
					members := decodeConfiguration(v[17:])
					if len(members) > 0 {
						clusterMembers = members
						fmt.Printf("      Cluster Members: %v\n", members)
					}
				}

				entryCount++
			}
			return nil
		})

		if entryCount == 0 {
			fmt.Println("    (no entries)")
		}

		if len(clusterMembers) > 0 {
			fmt.Printf("\n  Current Cluster Configuration:\n")
			fmt.Printf("    Total Members: %d\n", len(clusterMembers))
			for i, member := range clusterMembers {
				fmt.Printf("    Member %d: %s\n", i+1, member)
			}
		}

		return nil
	})
}

func getLogTypeName(logType uint8) string {
	switch raft.LogType(logType) {
	case raft.LogCommand:
		return "Command"
	case raft.LogConfiguration:
		return "Configuration"
	case raft.LogAddPeerDeprecated:
		return "AddPeer (deprecated)"
	case raft.LogRemovePeerDeprecated:
		return "RemovePeer (deprecated)"
	case raft.LogBarrier:
		return "Barrier"
	default:
		return fmt.Sprintf("Unknown(%d)", logType)
	}
}

func determineRole(votedFor string) string {
	if votedFor != "" {
		return "Leader (voted for self)"
	}
	return "Follower"
}

func decodeConfiguration(data []byte) []string {
	if len(data) < 1 {
		return nil
	}

	// Skip protocol version
	pos := 1
	var members []string

	for pos < len(data) {
		// Read suffrage (1 byte)
		if pos >= len(data) {
			break
		}
		suffrage := data[pos]
		pos++

		// Read ID length
		if pos+8 > len(data) {
			break
		}
		idLen := binary.BigEndian.Uint64(data[pos : pos+8])
		pos += 8

		// Read ID
		if pos+int(idLen) > len(data) {
			break
		}
		id := string(data[pos : pos+int(idLen)])
		pos += int(idLen)

		// Read address length
		if pos+8 > len(data) {
			break
		}
		addrLen := binary.BigEndian.Uint64(data[pos : pos+8])
		pos += 8

		// Read address
		if pos+int(addrLen) > len(data) {
			break
		}
		addr := string(data[pos : pos+int(addrLen)])
		pos += int(addrLen)

		suffrageName := "Unknown"
		if suffrage == 0 {
			suffrageName = "Voter"
		} else if suffrage == 1 {
			suffrageName = "NonVoter"
		}

		members = append(members, fmt.Sprintf("%s (%s) - %s", id, addr, suffrageName))
	}

	return members
}
