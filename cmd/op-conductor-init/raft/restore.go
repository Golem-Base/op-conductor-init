package raft

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

// RestoreAction handles the restore subcommand
func RestoreAction(ctx *cli.Context) error {
	backupDir := ctx.String("backup-dir")
	stateDir := ctx.String("state-dir")
	force := ctx.Bool("force")

	fmt.Printf("Restoring Raft state from backup...\n")
	fmt.Printf("Source: %s\n", backupDir)
	fmt.Printf("Destination: %s\n", stateDir)

	// Check if backup directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return fmt.Errorf("backup directory does not exist: %s", backupDir)
	}

	// Check for metadata file
	metadataPath := filepath.Join(backupDir, "backup-metadata.txt")
	if _, err := os.Stat(metadataPath); err == nil {
		// Read and display metadata
		content, err := os.ReadFile(metadataPath)
		if err == nil {
			fmt.Printf("\nBackup metadata:\n")
			fmt.Printf("%s\n", string(content))
		}
	}

	// Find all .db files in backup directory
	var filesToRestore []string
	err := filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".db" {
			filesToRestore = append(filesToRestore, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to scan backup directory: %w", err)
	}

	if len(filesToRestore) == 0 {
		return fmt.Errorf("no backup files found in %s", backupDir)
	}

	// Check for existing files if not forcing
	if !force {
		var existingFiles []string
		for _, srcPath := range filesToRestore {
			relPath, err := filepath.Rel(backupDir, srcPath)
			if err != nil {
				continue
			}
			dstPath := filepath.Join(stateDir, relPath)
			if _, err := os.Stat(dstPath); err == nil {
				existingFiles = append(existingFiles, dstPath)
			}
		}

		if len(existingFiles) > 0 {
			fmt.Printf("\nWarning: The following files will be overwritten:\n")
			for _, file := range existingFiles {
				fmt.Printf("  - %s\n", file)
			}

			if !promptConfirmation("\nContinue with restore?") {
				fmt.Println("Restore cancelled.")
				return nil
			}
		}
	}

	// Create state directory if it doesn't exist
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Restore each file
	fmt.Printf("\nRestoring %d files:\n", len(filesToRestore))
	for _, srcPath := range filesToRestore {
		// Calculate relative path from backup directory
		relPath, err := filepath.Rel(backupDir, srcPath)
		if err != nil {
			return fmt.Errorf("failed to calculate relative path: %w", err)
		}

		// Skip metadata file
		if strings.HasSuffix(relPath, "backup-metadata.txt") {
			continue
		}

		// Create destination path
		dstPath := filepath.Join(stateDir, relPath)

		// Create destination directory if needed
		dstDir := filepath.Dir(dstPath)
		if err := os.MkdirAll(dstDir, 0o755); err != nil {
			return fmt.Errorf("failed to create destination directory: %w", err)
		}

		// Copy file
		if err := copyFile(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to restore %s: %w", relPath, err)
		}

		fmt.Printf("  ✓ %s\n", relPath)
	}

	fmt.Printf("\n✓ Restore completed successfully\n")
	fmt.Printf("State restored to: %s\n", stateDir)

	return nil
}

// promptConfirmation asks the user for yes/no confirmation
func promptConfirmation(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", prompt)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
