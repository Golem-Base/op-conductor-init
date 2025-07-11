package raft

import (
	"github.com/urfave/cli/v2"

	"github.com/ethereum-optimism/optimism/op-service/cliapp"
	"github.com/golem-base/op-conductor-init/pkg/flags"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:        "raft",
		Usage:       "Raft state management commands",
		Description: "Commands for initializing and managing Raft state files for op-conductor high-availability clusters",
		Subcommands: []*cli.Command{
			{
				Name:        "generate",
				Usage:       "Generate pre-configured Raft state",
				Description: "Generate Raft state files for all nodes in the cluster",
				Action:      GenerateAction,
				Flags:       cliapp.ProtectFlags(flags.Flags),
			},
			{
				Name:        "verify",
				Usage:       "Verify generated Raft state",
				Description: "Inspect and verify the contents of generated Raft state files",
				Action:      VerifyAction,
				Flags: cliapp.ProtectFlags([]cli.Flag{
					flags.StateDirFlag,
				}),
			},
			{
				Name:        "info",
				Usage:       "Show detailed information about Raft state",
				Description: "Display comprehensive information about the Raft state including term, leader, and configuration",
				Action:      InfoAction,
				Flags: cliapp.ProtectFlags([]cli.Flag{
					flags.StateDirFlag,
				}),
			},
			{
				Name:        "backup",
				Usage:       "Backup Raft state files",
				Description: "Create a timestamped backup of Raft state files",
				Action:      BackupAction,
				Flags: cliapp.ProtectFlags([]cli.Flag{
					flags.StateDirFlag,
					flags.BackupDirFlag,
				}),
			},
			{
				Name:        "restore",
				Usage:       "Restore Raft state from backup",
				Description: "Restore Raft state files from a previous backup",
				Action:      RestoreAction,
				Flags: cliapp.ProtectFlags([]cli.Flag{
					flags.BackupDirFlag,
					flags.StateDirFlag,
					flags.RestoreForceFlag,
				}),
			},
		},
	}
}
