package raft

import (
	"encoding/binary"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/raft"
	"github.com/urfave/cli/v2"
	bolt "go.etcd.io/bbolt"
)

// VerifyAction handles the verify subcommand
func VerifyAction(ctx *cli.Context) error {
	stateDir := ctx.String("state-dir")

	fmt.Printf("Verifying Raft state in: %s\n\n", stateDir)

	// Check stable store
	fmt.Println("=== Stable Store (raft-stable.db) ===")
	stablePath := filepath.Join(stateDir, "raft-stable.db")
	if err := verifyStableStore(stablePath); err != nil {
		return fmt.Errorf("error verifying stable store: %w", err)
	}

	fmt.Println("\n=== Log Store (raft-log.db) ===")
	logPath := filepath.Join(stateDir, "raft-log.db")
	if err := verifyLogStore(logPath); err != nil {
		return fmt.Errorf("error verifying log store: %w", err)
	}

	return nil
}

func verifyStableStore(path string) error {
	db, err := bolt.Open(path, 0o600, nil)
	if err != nil {
		return fmt.Errorf("failed to open stable store: %w", err)
	}
	defer db.Close()

	return db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("conf"))
		if bucket == nil {
			return fmt.Errorf("conf bucket not found")
		}

		// Print all key-value pairs
		return bucket.ForEach(func(k, v []byte) error {
			key := string(k)
			switch key {
			case "CurrentTerm", "LastVoteTerm":
				if len(v) == 8 {
					value := binary.BigEndian.Uint64(v)
					fmt.Printf("  %s: %d\n", key, value)
				} else {
					fmt.Printf("  %s: %x (invalid length: %d)\n", key, v, len(v))
				}
			case "LastVoteCand":
				fmt.Printf("  %s: %s\n", key, string(v))
			default:
				fmt.Printf("  %s: %x\n", key, v)
			}
			return nil
		})
	})
}

func verifyLogStore(path string) error {
	db, err := bolt.Open(path, 0o600, nil)
	if err != nil {
		return fmt.Errorf("failed to open log store: %w", err)
	}
	defer db.Close()

	return db.View(func(tx *bolt.Tx) error {
		// Check conf bucket
		confBucket := tx.Bucket([]byte("conf"))
		if confBucket != nil {
			fmt.Println("  Conf bucket:")
			confBucket.ForEach(func(k, v []byte) error {
				if string(k) == "FirstIndex" || string(k) == "LastIndex" {
					if len(v) == 8 {
						fmt.Printf("    %s: %d\n", k, binary.BigEndian.Uint64(v))
					}
				}
				return nil
			})
		}

		// Check logs bucket
		logsBucket := tx.Bucket([]byte("logs"))
		if logsBucket == nil {
			return fmt.Errorf("logs bucket not found")
		}

		fmt.Println("  Logs bucket:")
		count := 0
		return logsBucket.ForEach(func(k, v []byte) error {
			if len(k) == 8 && len(v) >= 17 {
				index := binary.BigEndian.Uint64(k)
				term := binary.BigEndian.Uint64(v[0:8])
				logIndex := binary.BigEndian.Uint64(v[8:16])
				logType := v[16]
				dataLen := len(v) - 17

				fmt.Printf("    Entry %d: term=%d, index=%d, type=%d, data_len=%d\n",
					index, term, logIndex, logType, dataLen)

				// If it's a configuration entry, try to decode it
				if logType == uint8(raft.LogConfiguration) {
					fmt.Printf("      Configuration data (first 50 bytes): %x\n", v[17:min(67, len(v))])
				}
			}
			count++
			return nil
		})
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
