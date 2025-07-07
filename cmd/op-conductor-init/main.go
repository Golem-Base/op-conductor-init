package main

import (
	"os"

	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"

	opservice "github.com/ethereum-optimism/optimism/op-service"
	"github.com/ethereum-optimism/optimism/op-service/cliapp"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/golem-base/op-conductor-init/pkg/flags"
)

var (
	Version   = "v0.0.1"
	GitCommit = ""
	GitDate   = ""
)

func main() {
	oplog.SetupDefaults()

	app := cli.NewApp()
	app.Version = opservice.FormatVersion(Version, GitCommit, GitDate, "")
	app.Name = "op-conductor-init"
	app.Usage = "Initialize Raft state for op-conductor clusters"
	app.Description = "Tool to initialize Raft state files for op-conductor high-availability clusters"
	app.Commands = []*cli.Command{
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
				&cli.StringFlag{
					Name:     "state-dir",
					Usage:    "Directory containing raft state files to verify",
					Required: true,
				},
			}),
		},
		{
			Name:        "info",
			Usage:       "Show detailed information about Raft state",
			Description: "Display comprehensive information about the Raft state including term, leader, and configuration",
			Action:      InfoAction,
			Flags: cliapp.ProtectFlags([]cli.Flag{
				&cli.StringFlag{
					Name:     "state-dir",
					Usage:    "Directory containing raft state files",
					Required: true,
				},
			}),
		},
		{
			Name:        "backup",
			Usage:       "Backup Raft state files",
			Description: "Create a timestamped backup of Raft state files",
			Action:      BackupAction,
			Flags: cliapp.ProtectFlags([]cli.Flag{
				&cli.StringFlag{
					Name:     "state-dir",
					Usage:    "Directory containing raft state files to backup",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "backup-dir",
					Usage:    "Directory where backup will be created",
					Required: true,
				},
			}),
		},
		{
			Name:        "restore",
			Usage:       "Restore Raft state from backup",
			Description: "Restore Raft state files from a previous backup",
			Action:      RestoreAction,
			Flags: cliapp.ProtectFlags([]cli.Flag{
				&cli.StringFlag{
					Name:     "backup-dir",
					Usage:    "Directory containing the backup to restore",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "state-dir",
					Usage:    "Directory where state will be restored",
					Required: true,
				},
				&cli.BoolFlag{
					Name:  "force",
					Usage: "Force restore without confirmation prompts",
					Value: false,
				},
			}),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Crit("Application failed", "message", err)
	}
}
