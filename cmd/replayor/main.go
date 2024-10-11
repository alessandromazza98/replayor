package main

import (
	"context"
	"os"

	"github.com/danyalprout/replayor/packages/clients"
	"github.com/danyalprout/replayor/packages/config"
	"github.com/danyalprout/replayor/packages/replayor"
	"github.com/danyalprout/replayor/packages/stats"
	opservice "github.com/ethereum-optimism/optimism/op-service"
	"github.com/ethereum-optimism/optimism/op-service/cliapp"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

var (
	Version   = "v0.0.1"
	GitCommit = ""
	GitDate   = ""
)

func main() {
	oplog.SetupDefaults()
	oplog.SetGlobalLogHandler(log.LogfmtHandlerWithLevel(os.Stdout, log.LevelTrace))
	app := cli.NewApp()
	app.Flags = cliapp.ProtectFlags(config.Flags)
	app.Version = opservice.FormatVersion(Version, GitCommit, GitDate, "")
	app.Name = "replayor"
	app.Description = "Utility to interact with the engine api and emit metrics for block building"
	app.Action = cliapp.LifecycleCmd(Main())

	err := app.Run(os.Args)
	if err != nil {
		log.Crit("Application failed", "message", err)
	}
}

func Main() cliapp.LifecycleAction {
	return func(cliCtx *cli.Context, close context.CancelCauseFunc) (cliapp.Lifecycle, error) {
		ctx := cliCtx.Context
		logger := oplog.NewLogger(oplog.AppOut(cliCtx), oplog.ReadCLIConfig(cliCtx))

		cfg, err := config.LoadReplayorConfig(cliCtx, logger)
		if err != nil {
			return nil, err
		}

		c, err := clients.SetupClients(cfg, logger, ctx)
		if err != nil {
			return nil, err
		}

		// Benchmark stats
		s, err := stats.NewStorage(logger, cfg)
		if err != nil {
			return nil, err
		}
		statsRecorder := stats.NewStoredStats(s, logger)

		return replayor.NewService(c, statsRecorder, cfg, logger, close), nil
	}
}
