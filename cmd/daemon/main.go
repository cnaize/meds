package main

import (
	"context"
	"flag"
	"fmt"
	"runtime"
	"time"

	"github.com/appleboy/graceful"
	"github.com/rs/zerolog"

	"github.com/cnaize/meds/lib/util/get"
	"github.com/cnaize/meds/src/config"
	"github.com/cnaize/meds/src/core"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/database"
	"github.com/cnaize/meds/src/server"
	"github.com/cnaize/meds/src/types"

	dnsfilter "github.com/cnaize/meds/src/core/filter/dns"
	ipfilter "github.com/cnaize/meds/src/core/filter/ip"
	ratefilter "github.com/cnaize/meds/src/core/filter/rate"
)

func main() {
	var cfg config.Config
	// parse config
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "zerolog level")
	flag.StringVar(&cfg.DBFilePath, "db-path", "meds.db", "path to database file")
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

	// create database
	db := database.NewDatabase(cfg.DBFilePath, logger)
	if err := db.Init(mainCtx); err != nil {
		logger.Raw().Fatal().Err(err).Msg("database init failed")
	}

	// load white/black lists
	subnetWhiteList, subnetBlackList, domainWhiteList, domainBlackList, err := loadWhiteBlackLists(mainCtx, db)
	if err != nil {
		logger.Raw().Fatal().Err(err).Msg("white/black lists load")
	}

	// create filters
	filters := newFilters(cfg, logger)

	// create queue
	q := core.NewQueue(cfg.WorkersCount, subnetWhiteList, subnetBlackList, domainWhiteList, domainBlackList, filters, logger)
	if err := q.Load(mainCtx); err != nil {
		logger.Raw().Fatal().Err(err).Msg("queue load failed")
	}
	go q.Update(mainCtx, cfg.UpdateTimeout, cfg.UpdateInterval)

	// create server
	api := server.NewServer(
		cfg.APIServerAddr, cfg.Username, cfg.Password, db, subnetWhiteList, subnetBlackList, domainWhiteList, domainBlackList,
	)

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

		// close database
		if err := db.Close(); err != nil {
			logger.Raw().Err(err).Msg("database close failed")
		}

		return nil
	})

	// wait till the end
	<-m.Done()
}

func loadWhiteBlackLists(ctx context.Context, db *database.Database) (*types.SubnetList, *types.SubnetList, *types.DomainList, *types.DomainList, error) {
	// load subnet whitelist
	subnetWhiteList := types.NewSubnetList()
	snWhiteList, err := db.Q.GetAllWhiteListSubnets(ctx, db.DB)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("subnet whitelist get: %w", err)
	}
	if len(snWhiteList) > 0 {
		subnets, err := get.Subnets(snWhiteList)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("subnet whitelist parse: %w", err)
		}
		if err := subnetWhiteList.Upsert(subnets); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("subnet whitelist upsert: %w", err)
		}
	} else {
		if err := prefillWhiteList(ctx, db, subnetWhiteList); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("subnet whitelist prefill: %w", err)
		}
	}

	// load subnet blacklist
	subnetBlackList := types.NewSubnetList()
	snBlackList, err := db.Q.GetAllBlackListSubnets(ctx, db.DB)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("subnet blacklist get: %w", err)
	}
	subnets, err := get.Subnets(snBlackList)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("subnet blacklist parse: %w", err)
	}
	if err := subnetBlackList.Upsert(subnets); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("subnet blacklist upsert: %w", err)
	}

	// load domain whitelist
	domainWhiteList := types.NewDomainList()
	dmWhiteList, err := db.Q.GetAllWhiteListDomains(ctx, db.DB)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("domain whitelist get: %w", err)
	}
	if err := domainWhiteList.Upsert(dmWhiteList); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("domain whitelist upsert: %w", err)
	}

	// load domain whitelist
	domainBlackList := types.NewDomainList()
	dmBlackList, err := db.Q.GetAllBlackListDomains(ctx, db.DB)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("domain blacklist get: %w", err)
	}
	if err := domainBlackList.Upsert(dmBlackList); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("domain blacklist upsert: %w", err)
	}

	return subnetWhiteList, subnetBlackList, domainWhiteList, domainBlackList, nil
}

func prefillWhiteList(ctx context.Context, db *database.Database, subnetWhiteList *types.SubnetList) error {
	prefill := []string{
		"127.0.0.0/8",
		"10.0.0.0/8",
		"192.168.0.0/16",
		"172.16.0.0/12",
	}

	subnets, err := get.Subnets(prefill)
	if err != nil {
		return fmt.Errorf("get subnets: %w", err)
	}

	if err := subnetWhiteList.Upsert(subnets); err != nil {
		return fmt.Errorf("upsert subnets: %w", err)
	}

	for _, subnet := range prefill {
		if err := db.Q.UpsertWhiteListSubnet(ctx, db.DB, subnet); err != nil {
			return fmt.Errorf("upsert subnet: %w", err)
		}
	}

	return nil
}

func newFilters(cfg config.Config, logger *logger.Logger) []filter.Filter {
	return []filter.Filter{
		// rate filter
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
}
