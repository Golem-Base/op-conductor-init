package raft

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v2"
)

// BackupAction handles the backup subcommand
func BackupAction(ctx *cli.Context) error {
	stateDir := ctx.String("state-dir")
	backupDir := ctx.String("backup-dir")

	// Create timestamp for backup
	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("raft-backup-%s", timestamp))

	fmt.Printf("Creating backup of Raft state...\n")
	fmt.Printf("Source: %s\n", stateDir)
	fmt.Printf("Destination: %s\n", backupPath)

	// Create backup directory
	if err := os.MkdirAll(backupPath, 0o755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Find all .db files in state directory
	var filesToBackup []string
	err := filepath.Walk(stateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".db" {
			filesToBackup = append(filesToBackup, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to scan state directory: %w", err)
	}

	if len(filesToBackup) == 0 {
		return fmt.Errorf("no state files found in %s", stateDir)
	}

	// Backup each file
	fmt.Printf("\nBacking up %d files:\n", len(filesToBackup))
	for _, srcPath := range filesToBackup {
		// Calculate relative path from state directory
		relPath, err := filepath.Rel(stateDir, srcPath)
		if err != nil {
			return fmt.Errorf("failed to calculate relative path: %w", err)
		}

		// Create destination path
		dstPath := filepath.Join(backupPath, relPath)

		// Create destination directory if needed
		dstDir := filepath.Dir(dstPath)
		if err := os.MkdirAll(dstDir, 0o755); err != nil {
			return fmt.Errorf("failed to create destination directory: %w", err)
		}

		// Copy file
		if err := copyFile(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to backup %s: %w", relPath, err)
		}

		fmt.Printf("  ✓ %s\n", relPath)
	}

	// Create metadata file
	metadataPath := filepath.Join(backupPath, "backup-metadata.txt")
	metadata := fmt.Sprintf("Backup created: %s\nSource directory: %s\nFiles backed up: %d\n",
		time.Now().Format(time.RFC3339),
		stateDir,
		len(filesToBackup))

	if err := os.WriteFile(metadataPath, []byte(metadata), 0o644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	fmt.Printf("\n✓ Backup completed successfully\n")
	fmt.Printf("Backup location: %s\n", backupPath)

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Get source file info
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	// Create destination file
	destFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy contents
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Sync to ensure data is written
	return destFile.Sync()
}
