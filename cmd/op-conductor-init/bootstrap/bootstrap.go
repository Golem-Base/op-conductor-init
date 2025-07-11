package bootstrap

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/ethereum-optimism/optimism/op-conductor/conductor"
	opcflags "github.com/ethereum-optimism/optimism/op-conductor/flags"
	opservice "github.com/ethereum-optimism/optimism/op-service"
	"github.com/ethereum-optimism/optimism/op-service/cliapp"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	bootstrapconductor "github.com/golem-base/op-conductor-init/pkg/conductor"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:        "bootstrap",
		Usage:       "Bootstrap commands for op-conductor initialization",
		Description: "Commands for bootstrapping op-conductor clusters and related operations",
		Subcommands: []*cli.Command{
			{
				Name:        "cluster",
				Usage:       "Bootstrap a new op-conductor cluster",
				Description: "Initialize and bootstrap a new op-conductor cluster with the specified configuration",
				Action:      cliapp.LifecycleCmd(BootstrapClusterMain),
				Flags:       cliapp.ProtectFlags(opcflags.Flags),
			},
		},
	}
}

// BootstrapClusterMain handles the bootstrap cluster subcommand with lifecycle management
func BootstrapClusterMain(ctx *cli.Context, closeApp context.CancelCauseFunc) (cliapp.Lifecycle, error) {
	logCfg := oplog.ReadCLIConfig(ctx)
	log := oplog.NewLogger(oplog.AppOut(ctx), logCfg)
	oplog.SetGlobalLogHandler(log.Handler())
	opservice.ValidateEnvVars(opcflags.EnvVarPrefix, opcflags.Flags, log)

	cfg, err := conductor.NewConfig(ctx, log)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// We don't need the RaftAutoBootstrapNetwork flag anymore
	// Our custom OpConductor always auto-bootstraps when there's no unsafe head
	log.Info("Bootstrap operation will auto-bootstrap from execution layer when needed")

	// Metrics are created internally by our custom OpConductor

	// Use our custom OpConductor from service.go that has auto-bootstrap built in
	c, err := bootstrapconductor.New(
		ctx.Context,
		cfg,
		log,
		"bootstrap-v1.0.0", // Version string for bootstrap
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create conductor: %w", err)
	}

	return c, nil
}
