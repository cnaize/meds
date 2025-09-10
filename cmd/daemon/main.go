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

	dnsfilter "github.com/cnaize/meds/src/core/filter/dns"
	ipfilter "github.com/cnaize/meds/src/core/filter/ip"
)

func main() {
	var cfg config.Config
	// parse config
	flag.UintVar(&cfg.QueueCount, "qcount", uint(runtime.GOMAXPROCS(0)), "set nfqueue count")
	flag.DurationVar(&cfg.UpdateTimeout, "update-timeout", time.Minute, "update timeout per filter")
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
		ipfilter.NewFireHOL([]string{
			"https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level1.netset",
			"https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level2.netset",
		}, logger),
		dnsfilter.NewStevenBlack([]string{
			"https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
		}, logger),
	}

	// create queue
	q := core.NewQueue(cfg.QueueCount, filters, logger)
	if err := q.Load(mainCtx); err != nil {
		panic(fmt.Sprintf("Failed to load queue: %s", err.Error()))
	}
	go q.Update(mainCtx, cfg.UpdateTimeout, cfg.UpdateInterval)

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
