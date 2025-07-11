package main

import (
	"context"
	"os"

	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"

	opservice "github.com/ethereum-optimism/optimism/op-service"
	"github.com/ethereum-optimism/optimism/op-service/ctxinterrupt"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/golem-base/op-conductor-init/cmd/op-conductor-init/bootstrap"
	"github.com/golem-base/op-conductor-init/cmd/op-conductor-init/raft"
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
		raft.Command(),
		bootstrap.Command(),
	}

	ctx := ctxinterrupt.WithSignalWaiterMain(context.Background())
	err := app.RunContext(ctx, os.Args)
	if err != nil {
		log.Crit("Application failed", "message", err)
	}
}
