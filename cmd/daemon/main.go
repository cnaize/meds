package main

import (
	"context"
	"flag"
	"runtime"

	"github.com/appleboy/graceful"

	"github.com/cnaize/meds/src/config"
	"github.com/cnaize/meds/src/core"
	"github.com/cnaize/meds/src/core/filter"
)

func main() {
	var cfg config.Config
	// parse config
	flag.UintVar(&cfg.QCount, "qcount", uint(runtime.GOMAXPROCS(0)), "set nfqueue count")
	flag.Parse()

	// create filters
	filters := []core.Filter{
		filter.NewTrue(),
	}

	// create logger
	logger := graceful.NewLogger()
	logger.Infof("Running Meds...")

	// main context
	mainCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create queue
	q := core.NewQueue(cfg.QCount, filters, logger)

	// run queue
	m := graceful.NewManager(graceful.WithContext(mainCtx), graceful.WithLogger(logger))
	m.AddRunningJob(func(ctx context.Context) error {
		defer cancel()

		select {
		case <-ctx.Done():
		default:
			if err := q.Run(ctx); err != nil {
				logger.Errorf("run: %+v", err)
			}
		}

		return nil
	})
	m.AddShutdownJob(q.Close)

	// wait till the end
	<-m.Done()
}
