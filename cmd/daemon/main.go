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
	"github.com/cnaize/meds/src/server"

	dnsfilter "github.com/cnaize/meds/src/core/filter/dns"
	ipfilter "github.com/cnaize/meds/src/core/filter/ip"
	ratefilter "github.com/cnaize/meds/src/core/filter/rate"
)

func main() {
	var cfg config.Config
	// parse config
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "zerolog level")
	flag.StringVar(&cfg.Username, "username", "admin", "admin username")
	flag.StringVar(&cfg.Password, "password", "admin", "admin password")
	flag.StringVar(&cfg.APIServerAddr, "api-addr", ":8000", "api server address")
	flag.UintVar(&cfg.WorkersCount, "workers-count", uint(runtime.GOMAXPROCS(0)), "nfqueue workers count")
	flag.UintVar(&cfg.LoggersCount, "loggers-count", uint(runtime.GOMAXPROCS(0)), "logger workers count")
	flag.DurationVar(&cfg.UpdateTimeout, "update-timeout", 10*time.Second, "update timeout (per filter)")
	flag.DurationVar(&cfg.UpdateInterval, "update-interval", 12*time.Hour, "update frequency")
	flag.UintVar(&cfg.LimiterMaxBalance, "max-packets-at-once", 2000, "max packets per ip at once")
	flag.UintVar(&cfg.LimiterRefillRate, "max-packets-per-second", 100, "max packets per ip per second")
	flag.UintVar(&cfg.LimiterCacheSize, "max-packets-cache-size", 10_000, "max packets per ip cache size")
	flag.DurationVar(&cfg.LimiterBucketTTL, "max-packets-cache-ttl", 3*time.Minute, "max packets per ip cache ttl")
	flag.Parse()

	// set "debug" for invalid log level
	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		logLevel = zerolog.DebugLevel
	}

	// main context
	mainCtx, mainCancel := context.WithCancel(context.Background())
	defer mainCancel()

	// create logger
	logger := logger.NewLogger(get.Ptr(
		zerolog.New(zerolog.NewConsoleWriter()).
			With().
			Timestamp().
			Logger().
			Level(logLevel),
	))
	logger.Run(mainCtx, cfg.LoggersCount)

	logger.Raw().Info().Msg("Running Meds...")

	// create filters
	filters := []filter.Filter{
		// rate filters
		ratefilter.NewLimiter(cfg.LimiterMaxBalance, cfg.LimiterRefillRate, cfg.LimiterCacheSize, cfg.LimiterBucketTTL, logger),
		// ip filters
		ipfilter.NewFireHOL([]string{
			"https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level1.netset",
			"https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level2.netset",
		}, logger),
		ipfilter.NewSpamhaus([]string{
			"https://www.spamhaus.org/drop/drop.txt",
		}, logger),
		ipfilter.NewAbuse([]string{
			"https://feodotracker.abuse.ch/downloads/ipblocklist.txt",
		}, logger),
		// dns filters
		dnsfilter.NewStevenBlack([]string{
			"https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
		}, logger),
		dnsfilter.NewSomeoneWhoCares([]string{
			"https://someonewhocares.org/hosts/hosts",
		}, logger),
	}

	// create queue
	q := core.NewQueue(cfg.WorkersCount, filters, logger)
	if err := q.Load(mainCtx); err != nil {
		logger.Raw().Fatal().Err(err).Msg("queue load failed")
	}
	go q.Update(mainCtx, cfg.UpdateTimeout, cfg.UpdateInterval)

	// create server
	api := server.NewServer(cfg.APIServerAddr, cfg.Username, cfg.Password)

	m := graceful.NewManager(graceful.WithContext(mainCtx), graceful.WithLogger(graceful.NewLogger()))
	m.AddRunningJob(func(ctx context.Context) error {
		defer mainCancel()

		// run server
		go func() {
			defer mainCancel()

			if err := api.Run(ctx); err != nil {
				logger.Raw().Err(err).Msg("api run failed")
			}
		}()

		// run queue
		if err := q.Run(ctx); err != nil {
			logger.Raw().Err(err).Msg("queue run failed")
		}

		return nil
	})
	m.AddShutdownJob(func() error {
		// close server
		if err := api.Close(); err != nil {
			logger.Raw().Err(err).Msg("api close failed")
		}

		// close queue
		if err := q.Close(); err != nil {
			logger.Raw().Err(err).Msg("queue close failed")
		}

		return nil
	})

	// wait till the end
	<-m.Done()
}
