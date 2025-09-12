package main

import (
	"context"
	"flag"
	"runtime"
	"time"

	"github.com/appleboy/graceful"
	"github.com/rs/zerolog"

	"github.com/cnaize/meds/lib/util/get"
	"github.com/cnaize/meds/src/config"
	"github.com/cnaize/meds/src/core"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"

	ipfilter "github.com/cnaize/meds/src/core/filter/ip"
)

func main() {
	var cfg config.Config
	// parse config
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "zerolog level")
	flag.UintVar(&cfg.WorkersCount, "workers-count", uint(runtime.GOMAXPROCS(0)), "nfqueue workers count")
	flag.UintVar(&cfg.LoggersCount, "loggers-count", uint(runtime.GOMAXPROCS(0)), "logger workers count")
	flag.DurationVar(&cfg.UpdateTimeout, "update-timeout", time.Minute, "update timeout (per filter)")
	flag.DurationVar(&cfg.UpdateInterval, "update-interval", 12*time.Hour, "update frequency")
	flag.Parse()

	// set "debug" for invalid log level
	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		logLevel = zerolog.DebugLevel
	}

	// create logger
	logger := logger.NewLogger(get.Ptr(
		zerolog.New(
			zerolog.NewConsoleWriter(),
		).
			With().
			Timestamp().
			Logger().
			Level(logLevel)),
	)
	logger.Run(cfg.LoggersCount)

	logger.Logger().Info().Msg("Running Meds...")

	// main context
	mainCtx, mainCancel := context.WithCancel(context.Background())
	defer mainCancel()

	// create filters
	filters := []filter.Filter{
		ipfilter.NewFireHOL([]string{"https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level1.netset"}, logger),
		ipfilter.NewSpamhaus([]string{"https://www.spamhaus.org/drop/drop.txt"}, logger),
		ipfilter.NewAbuse([]string{"https://feodotracker.abuse.ch/downloads/ipblocklist.txt"}, logger),
	}

	// create queue
	q := core.NewQueue(cfg.WorkersCount, filters, logger)
	if err := q.Load(mainCtx); err != nil {
		logger.Logger().Fatal().Err(err).Msg("queue load failed")
	}
	go q.Update(mainCtx, cfg.UpdateTimeout, cfg.UpdateInterval)

	// run queue
	m := graceful.NewManager(graceful.WithContext(mainCtx), graceful.WithLogger(graceful.NewLogger()))
	m.AddRunningJob(func(ctx context.Context) error {
		defer mainCancel()

		select {
		case <-ctx.Done():
		default:
			if err := q.Run(ctx); err != nil {
				logger.Logger().Err(err).Msg("queue run failed")
			}
		}

		return nil
	})
	m.AddShutdownJob(q.Close)

	// wait till the end
	<-m.Done()
}
