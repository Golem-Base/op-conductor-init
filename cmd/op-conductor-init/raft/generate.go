package raft

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	opservice "github.com/ethereum-optimism/optimism/op-service"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/golem-base/op-conductor-init/pkg/config"
	"github.com/golem-base/op-conductor-init/pkg/flags"
	"github.com/golem-base/op-conductor-init/pkg/generator"
)

// GenerateAction handles the generate subcommand
func GenerateAction(ctx *cli.Context) error {
	logCfg := oplog.ReadCLIConfig(ctx)
	log := oplog.NewLogger(oplog.AppOut(ctx), logCfg)
	oplog.SetGlobalLogHandler(log.Handler())
	opservice.ValidateEnvVars(flags.EnvVarPrefix, flags.Flags, log)

	cfg, err := config.NewConfig(ctx, log)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	gen := generator.New(cfg, log)
	if err := gen.Generate(context.Background()); err != nil {
		return fmt.Errorf("failed to generate raft state: %w", err)
	}

	log.Info("Successfully generated Raft state", "output", cfg.OutputDir)
	return nil
}
