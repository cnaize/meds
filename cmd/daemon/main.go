package main

import (
	"context"
	"flag"
	"fmt"
	"net/netip"
	"os"
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

	asnfilter "github.com/cnaize/meds/src/core/filter/asn"
	domainfilter "github.com/cnaize/meds/src/core/filter/domain"
	geofilter "github.com/cnaize/meds/src/core/filter/geo"
	ipfilter "github.com/cnaize/meds/src/core/filter/ip"
	ja3filter "github.com/cnaize/meds/src/core/filter/ja3"
	ratefilter "github.com/cnaize/meds/src/core/filter/rate"
)

func main() {
	var cfg config.Config
	// parse config
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "zerolog level")
	flag.StringVar(&cfg.DBFilePath, "db-path", "meds.db", "path to database file")
	flag.StringVar(&cfg.APIServerAddr, "api-addr", ":8000", "api server address")
	flag.UintVar(&cfg.ReadersCount, "readers-count", uint(runtime.GOMAXPROCS(0)), "nfqueue readers count")
	flag.UintVar(&cfg.WorkersCount, "workers-count", 1, "nfqueue workers count (per reader)")
	flag.UintVar(&cfg.LoggersCount, "loggers-count", uint(max(1, runtime.GOMAXPROCS(0)/4)), "logger workers count")
	flag.UintVar(&cfg.ReaderQLen, "reader-queue-len", 8192, "nfqueue queue length (per reader)")
	flag.UintVar(&cfg.LoggerQLen, "logger-queue-len", 2048, "logger queue length (all workers)")
	flag.DurationVar(&cfg.UpdateTimeout, "update-timeout", time.Minute, "update timeout (per filter)")
	flag.DurationVar(&cfg.UpdateInterval, "update-interval", 4*time.Hour, "update frequency")
	flag.UintVar(&cfg.LimiterRate, "rate-limiter-rate", 3000, "max packets per second (per ip)")
	flag.UintVar(&cfg.LimiterBurst, "rate-limiter-burst", 1500, "max packets at once (per ip)")
	flag.UintVar(&cfg.LimiterCacheSize, "rate-limiter-cache-size", 100_000, "rate limiter cache size (all buckets)")
	flag.DurationVar(&cfg.LimiterBucketTTL, "rate-limiter-cache-ttl", 3*time.Minute, "rate limiter cache ttl (per bucket)")
	// NOTE: set using "MEDS_USERNAME" and "MEDS_PASSWORD" environment variables
	// flag.StringVar(&cfg.Username, "username", "admin", "admin username")
	// flag.StringVar(&cfg.Password, "password", "admin", "admin password")
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
			Level(logLevel)),
		cfg.LoggerQLen,
	)
	logger.Run(mainCtx, cfg.LoggersCount)

	// check username/password
	cfg.Username = os.Getenv("MEDS_USERNAME")
	cfg.Password = os.Getenv("MEDS_PASSWORD")
	if len(cfg.Username) < 1 || len(cfg.Password) < 1 {
		logger.Raw().Fatal().Msg(`Please set "MEDS_USERNAME" and "MEDS_PASSWORD" environment variables`)
	}

	logger.Raw().Info().Msg("Running Meds...")

	// create database
	db := database.NewDatabase(cfg.DBFilePath, logger)
	if err := db.Init(mainCtx); err != nil {
		logger.Raw().Fatal().Err(err).Msg("database init failed")
	}

	// load white/black lists
	subnetWhiteList, subnetBlackList, domainWhiteList, domainBlackList, countryBlackList, err := loadWhiteBlackLists(mainCtx, db)
	if err != nil {
		logger.Raw().Fatal().Err(err).Msg("white/black lists load")
	}

	// create filters
	filters := newFilters(
		cfg,
		logger,
		subnetWhiteList,
		subnetBlackList,
		domainWhiteList,
		domainBlackList,
		countryBlackList,
	)

	// create queue
	q := core.NewQueue(cfg.ReadersCount, cfg.WorkersCount, cfg.ReaderQLen, filters, logger)
	if err := q.Load(mainCtx); err != nil {
		logger.Raw().Fatal().Err(err).Msg("queue load failed")
	}
	go q.Update(mainCtx, cfg.UpdateTimeout, cfg.UpdateInterval)

	// create server
	api := server.NewServer(
		cfg.APIServerAddr,
		cfg.Username,
		cfg.Password,
		db,
		subnetWhiteList,
		subnetBlackList,
		domainWhiteList,
		domainBlackList,
		countryBlackList,
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

func loadWhiteBlackLists(ctx context.Context, db *database.Database) (
	*types.SubnetList,
	*types.SubnetList,
	*types.DomainList,
	*types.DomainList,
	*types.CountryList,
	error,
) {
	// load subnet whitelist
	subnetWhiteList := types.NewSubnetList()
	snWhiteList, err := db.Q.GetAllWhiteListSubnets(ctx, db.DB)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("subnet whitelist get: %w", err)
	}
	if len(snWhiteList) > 0 {
		subnets, err := get.Subnets(snWhiteList)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("subnet whitelist parse: %w", err)
		}
		if err := subnetWhiteList.Upsert(subnets); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("subnet whitelist upsert: %w", err)
		}
	} else {
		if err := prefillWhiteList(ctx, db, subnetWhiteList); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("subnet whitelist prefill: %w", err)
		}
	}

	// load subnet blacklist
	subnetBlackList := types.NewSubnetList()
	snBlackList, err := db.Q.GetAllBlackListSubnets(ctx, db.DB)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("subnet blacklist get: %w", err)
	}
	subnets, err := get.Subnets(snBlackList)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("subnet blacklist parse: %w", err)
	}
	if err := subnetBlackList.Upsert(subnets); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("subnet blacklist upsert: %w", err)
	}

	// load domain whitelist
	domainWhiteList := types.NewDomainList()
	dmWhiteList, err := db.Q.GetAllWhiteListDomains(ctx, db.DB)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("domain whitelist get: %w", err)
	}
	if err := domainWhiteList.Upsert(dmWhiteList); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("domain whitelist upsert: %w", err)
	}

	// load domain whitelist
	domainBlackList := types.NewDomainList()
	dmBlackList, err := db.Q.GetAllBlackListDomains(ctx, db.DB)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("domain blacklist get: %w", err)
	}
	if err := domainBlackList.Upsert(dmBlackList); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("domain blacklist upsert: %w", err)
	}

	countryBlackList := types.NewCountryList()
	crBlackList, err := db.Q.GetAllBlackListCountries(ctx, db.DB)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("country blacklist get: %w", err)
	}
	if err := countryBlackList.Upsert(crBlackList); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("country blacklist upsert: %w", err)
	}

	return subnetWhiteList, subnetBlackList, domainWhiteList, domainBlackList, countryBlackList, nil
}

func prefillWhiteList(ctx context.Context, db *database.Database, subnetWhiteList *types.SubnetList) error {
	subnets := []netip.Prefix{
		netip.MustParsePrefix("127.0.0.0/8"),
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("172.16.0.0/12"),
	}

	// upsert to whitelist
	if err := subnetWhiteList.Upsert(subnets); err != nil {
		return fmt.Errorf("upsert subnets: %w", err)
	}

	// upsert to database
	for _, subnet := range subnets {
		if err := db.Q.UpsertWhiteListSubnet(ctx, db.DB, subnet.String()); err != nil {
			return fmt.Errorf("upsert subnet: %w", err)
		}
	}

	return nil
}

func newFilters(
	cfg config.Config,
	logger *logger.Logger,
	subnetWhiteList *types.SubnetList,
	subnetBlackList *types.SubnetList,
	domainWhiteList *types.DomainList,
	domainBlackList *types.DomainList,
	countryBlacklist *types.CountryList,
) []filter.Filter {
	// geofilter.IPLocate is responsible for the ASNList updates
	asnList := types.NewASNList()

	return []filter.Filter{
		// ip whitelist
		ipfilter.NewWhiteList(logger, subnetWhiteList),
		// rate filter
		ratefilter.NewLimiter(cfg.LimiterRate, cfg.LimiterBurst, cfg.LimiterCacheSize, cfg.LimiterBucketTTL, logger),
		// ip blacklist
		ipfilter.NewBlackList(logger, subnetBlackList),
		// ip filters
		ipfilter.NewFireHOL([]string{
			"https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level1.netset",
		}, logger),
		ipfilter.NewSpamhaus([]string{
			"https://www.spamhaus.org/drop/drop.txt",
		}, logger),
		ipfilter.NewAbuse([]string{
			"https://feodotracker.abuse.ch/downloads/ipblocklist.txt",
		}, logger),
		// geo filters
		geofilter.NewIPLocate([]string{
			"https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-asn/ip-to-asn.csv.zip",
		}, logger, asnList, countryBlacklist),
		// asn filters
		asnfilter.NewSpamhaus([]string{
			"https://www.spamhaus.org/drop/asndrop.json",
		}, logger, asnList),
		// domain/sni whitelist
		domainfilter.NewWhiteList(logger, domainWhiteList),
		// domain/sni blacklist
		domainfilter.NewBlackList(logger, domainBlackList),
		// domain/sni filters
		domainfilter.NewStevenBlack([]string{
			"https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
		}, logger),
		domainfilter.NewSomeoneWhoCares([]string{
			"https://someonewhocares.org/hosts/hosts",
		}, logger),
		// ja3 filters
		ja3filter.NewAbuse([]string{
			"https://sslbl.abuse.ch/blacklist/ja3_fingerprints.csv",
		}, logger),
	}
}
