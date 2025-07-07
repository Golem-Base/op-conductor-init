package flags

import (
	"github.com/urfave/cli/v2"

	opservice "github.com/ethereum-optimism/optimism/op-service"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
)

const EnvVarPrefix = "OP_CONDUCTOR_INIT"

var (
	NodesFlag = &cli.StringFlag{
		Name:     "nodes",
		Usage:    "Comma-separated list of node addresses (e.g., sequencer-1:50050,sequencer-2:50050,sequencer-3:50050)",
		EnvVars:  opservice.PrefixEnvVar(EnvVarPrefix, "NODES"),
		Required: true,
	}
	ServerIDsFlag = &cli.StringFlag{
		Name:     "server-ids",
		Usage:    "Comma-separated list of server IDs (e.g., sequencer-1,sequencer-2,sequencer-3)",
		EnvVars:  opservice.PrefixEnvVar(EnvVarPrefix, "SERVER_IDS"),
		Required: true,
	}
	OutputDirFlag = &cli.StringFlag{
		Name:    "output-dir",
		Usage:   "Output directory for generated Raft state",
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "OUTPUT_DIR"),
		Value:   "./raft-state",
	}
	InitialLeaderFlag = &cli.StringFlag{
		Name:     "initial-leader",
		Usage:    "Server ID of the initial leader",
		EnvVars:  opservice.PrefixEnvVar(EnvVarPrefix, "INITIAL_LEADER"),
		Required: true,
	}
	InitialTermFlag = &cli.Uint64Flag{
		Name:    "initial-term",
		Usage:   "Initial Raft term",
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "INITIAL_TERM"),
		Value:   1,
	}
	NetworkFlag = &cli.StringFlag{
		Name:    "network",
		Usage:   "Network name for configuration (e.g., base-mainnet, op-mainnet)",
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "NETWORK"),
	}
	ForceFlag = &cli.BoolFlag{
		Name:    "force",
		Usage:   "Force overwrite existing state files",
		Value:   false,
		EnvVars: opservice.PrefixEnvVar(EnvVarPrefix, "FORCE"),
	}
)

var Flags = []cli.Flag{
	NodesFlag,
	ServerIDsFlag,
	OutputDirFlag,
	InitialLeaderFlag,
	InitialTermFlag,
	NetworkFlag,
	ForceFlag,
}

func init() {
	Flags = append(Flags, oplog.CLIFlags(EnvVarPrefix)...)
}
