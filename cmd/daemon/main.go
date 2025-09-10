package main

import (
	"context"
	"flag"
	"fmt"
	"runtime"
	"time"

	"github.com/appleboy/graceful"

	"github.com/cnaize/meds/src/config"
	"github.com/cnaize/meds/src/core"
	"github.com/cnaize/meds/src/core/filter"
	ipfilter "github.com/cnaize/meds/src/core/filter/ip"
)

func main() {
	var cfg config.Config
	// parse config
	flag.UintVar(&cfg.QCount, "qcount", uint(runtime.GOMAXPROCS(0)), "set nfqueue count")
	flag.DurationVar(&cfg.UpdateInterval, "update-interval", 24*time.Hour, "update frequency")
	flag.Parse()

	// create logger
	logger := graceful.NewLogger()
	logger.Infof("Running Meds...")

	// main context
	mainCtx, mainCancel := context.WithCancel(context.Background())
	defer mainCancel()

	// create filters
	filters := []filter.Filter{
		ipfilter.NewFireHOL(logger),
	}

	// create queue
	q := core.NewQueue(cfg.QCount, filters, logger)
	if err := q.Load(mainCtx); err != nil {
		panic(fmt.Sprintf("Failed to load queue: %s", err.Error()))
	}
	go q.Update(mainCtx, cfg.UpdateInterval)

	// run queue
	m := graceful.NewManager(graceful.WithContext(mainCtx), graceful.WithLogger(logger))
	m.AddRunningJob(func(ctx context.Context) error {
		defer mainCancel()

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
