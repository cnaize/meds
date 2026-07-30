package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"net/netip"
	"os"
	"runtime"
	"time"

	"github.com/appleboy/graceful"
	nserver "github.com/nats-io/nats-server/v2/server"
	nclient "github.com/nats-io/nats.go"
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
	quarantinefilter "github.com/cnaize/meds/src/core/filter/quarantine"
	ratefilter "github.com/cnaize/meds/src/core/filter/rate"
)

var defaultAllowList = []netip.Prefix{
	netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("169.254.0.0/16"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("224.0.0.0/4"),
	netip.MustParsePrefix("240.0.0.0/4"),
	// for local PC only, otherwise remove it using API
	netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.168.0.0/16"),
}

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
	flag.BoolVar(&cfg.NatsEnable, "nats-enable", false, "enable nats server")
	flag.StringVar(&cfg.NatsHost, "nats-host", "localhost", "nats server host")
	flag.IntVar(&cfg.NatsPort, "nats-port", 4222, "nats server port")
	flag.UintVar(&cfg.QuarantineIPCacheSize, "quarantine-ip-cache-size", 100_000, "quarantine ip cache size (all entities)")
	flag.DurationVar(&cfg.QuarantineIPEntityTTL, "quarantine-ip-entity-ttl", 15*time.Minute, "quarantine ip cache ttl (per entity)")
	flag.UintVar(&cfg.LimiterRate, "rate-limiter-rate", 3000, "max packets per second (per ip)")
	flag.UintVar(&cfg.LimiterBurst, "rate-limiter-burst", 1500, "max packets at once (per ip)")
	flag.UintVar(&cfg.LimiterCacheSize, "rate-limiter-cache-size", 100_000, "rate limiter cache size (all buckets)")
	flag.DurationVar(&cfg.LimiterBucketTTL, "rate-limiter-bucket-ttl", 5*time.Minute, "rate limiter cache ttl (per bucket)")
	flag.BoolVar(&cfg.FilterAbuseIPDBEnable, "filter-abuseipdb-enable", false, "enable abuseipdb filter")
	flag.IntVar(&cfg.FilterAbuseIPDBConfidence, "filter-abuseipdb-confidence", 100, "abuseipdb filter minimum confidence")
	flag.BoolVar(&cfg.FilterAbuseIPDBReportAddr, "filter-abuseipdb-report-addr", false, "report quarantine addresses to abuseipdb")

	// NOTE: set using "MEDS_USERNAME" and "MEDS_PASSWORD" environment variables
	// flag.StringVar(&cfg.Username, "username", "admin", "admin username")
	// flag.StringVar(&cfg.Password, "password", "admin", "admin password")

	// NOTE: set using "MEDS_NATS_USERNAME" and "MEDS_NATS_PASSWORD" environment variables
	// flag.StringVar(&cfg.NatsUsername, "nats-username", "nats", "nats username")
	// flag.StringVar(&cfg.NatsPassword, "nats-password", "nats", "nats password")

	// NOTE: set using "MEDS_ABUSEIPDB_API_KEY" environment variable
	// flag.StringVar(&cfg.FilterAbuseIPDBApiKey, "filter-abuseipdb-api-key", "your_abuseipdb_token", "abuseipdb filter api key")

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
	logger := logger.NewLogger(
		new(
			zerolog.New(zerolog.NewConsoleWriter()).
				With().
				Timestamp().
				Logger().
				Level(logLevel),
		),
		cfg.LoggerQLen,
	)
	logger.Run(mainCtx, cfg.LoggersCount)

	// check username/password
	cfg.Username = os.Getenv("MEDS_USERNAME")
	cfg.Password = os.Getenv("MEDS_PASSWORD")
	if len(cfg.Username) < 1 || len(cfg.Password) < 1 {
		logger.Raw().Fatal().Msg(`Please set "MEDS_USERNAME" and "MEDS_PASSWORD" environment variables`)
	}

	// check nats username/password
	if cfg.NatsEnable {
		cfg.NatsUsername = os.Getenv("MEDS_NATS_USERNAME")
		cfg.NatsPassword = os.Getenv("MEDS_NATS_PASSWORD")
		if len(cfg.NatsUsername) < 1 || len(cfg.NatsPassword) < 1 {
			logger.Raw().Fatal().Msg(`Please set "MEDS_NATS_USERNAME" and "MEDS_NATS_PASSWORD" environment variables`)
		}
	}

	// check abuseipdb api key
	if cfg.FilterAbuseIPDBEnable {
		cfg.FilterAbuseIPDBApiKey = os.Getenv("MEDS_ABUSEIPDB_API_KEY")
		if len(cfg.FilterAbuseIPDBApiKey) < 1 {
			logger.Raw().Fatal().Msg(`Please set "MEDS_ABUSEIPDB_API_KEY" environment variable`)
		}
	}

	logger.Raw().Info().Msg("Running Meds...")

	// init database
	db, err := initDatabase(mainCtx, &cfg, logger)
	if err != nil {
		logger.Raw().Fatal().Err(err).Msg("database init failed")
	}

	// load allow/block lists
	ipAllowList, countryBlockList, err := loadAllowBlockLists(mainCtx, db)
	if err != nil {
		logger.Raw().Fatal().Err(err).Msg("allow/block lists load failed")
	}

	// load ip lists
	ipIncludeList, ipExcludeList, err := loadIPLists(mainCtx, db)
	if err != nil {
		logger.Raw().Fatal().Err(err).Msg("ip lists load failed")
	}

	// load asn lists
	asnIncludeList, asnExcludeList, err := loadASNLists(mainCtx, db)
	if err != nil {
		logger.Raw().Fatal().Err(err).Msg("asn lists load failed")
	}

	// load ja3 lists
	ja3IncludeList, ja3ExcludeList, err := loadJA3Lists(mainCtx, db)
	if err != nil {
		logger.Raw().Fatal().Err(err).Msg("ja3 lists load failed")
	}

	// load domain lists
	domainIncludeList, domainExcludeList, err := loadDomainLists(mainCtx, db)
	if err != nil {
		logger.Raw().Fatal().Err(err).Msg("domain lists load failed")
	}

	// create nats server
	natsServer, err := nserver.NewServer(&nserver.Options{
		Host:       cfg.NatsHost,
		Port:       cfg.NatsPort,
		Username:   cfg.NatsUsername,
		Password:   cfg.NatsPassword,
		DontListen: !cfg.NatsEnable,
	})
	if err != nil {
		logger.Raw().Fatal().Err(err).Msg("nats server start failed")
	}

	// start nats server
	go natsServer.Start()
	if !natsServer.ReadyForConnections(5 * time.Second) {
		logger.Raw().Fatal().Err(err).Msg("nats server not ready")
	}

	// create nats client
	natsClient, err := nclient.Connect("", nclient.InProcessServer(natsServer), nclient.UserInfo(cfg.NatsUsername, cfg.NatsPassword))
	if err != nil {
		logger.Raw().Fatal().Err(err).Msg("nats client connect failed")
	}

	// create filters
	filters := newFilters(
		&cfg,
		logger,
		natsClient,
		ipAllowList,
		countryBlockList,
		ipIncludeList,
		ipExcludeList,
		asnIncludeList,
		asnExcludeList,
		ja3IncludeList,
		ja3ExcludeList,
		domainIncludeList,
		domainExcludeList,
	)

	// create queue
	q := core.NewQueue(&cfg, filters, logger)
	if err := q.Load(mainCtx); err != nil {
		logger.Raw().Fatal().Err(err).Msg("queue load failed")
	}
	go q.Update(mainCtx, cfg.UpdateTimeout, cfg.UpdateInterval)

	// create server
	api := server.NewServer(
		&cfg,
		db,
		ipAllowList,
		countryBlockList,
		ipIncludeList,
		ipExcludeList,
		asnIncludeList,
		asnExcludeList,
		ja3IncludeList,
		ja3ExcludeList,
		domainIncludeList,
		domainExcludeList,
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

		// stop nats
		if err := natsClient.Drain(); err != nil {
			logger.Raw().Err(err).Msg("nats client drain failed")
		}
		natsServer.Shutdown()

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

func initDatabase(ctx context.Context, cfg *config.Config, logger *logger.Logger) (*database.Database, error) {
	// check database is new
	_, err := os.Stat(cfg.DBFilePath)
	isNewDatabase := errors.Is(err, fs.ErrNotExist)

	// init database
	db := database.NewDatabase(cfg, logger)
	if err := db.Init(ctx); err != nil {
		return nil, fmt.Errorf("init: %w", err)
	}

	// prefill database
	if isNewDatabase {
		for _, subnet := range defaultAllowList {
			if err := db.Q.UpsertIPAllowList(ctx, db.DB, subnet.String()); err != nil {
				return nil, fmt.Errorf("prefill ip allowlist: %w", err)
			}
		}
	}

	return db, nil
}

func newFilters(
	cfg *config.Config,
	logger *logger.Logger,
	natsClient *nclient.Conn,
	ipAllowList *types.IPList,
	countryBlocklist *types.CountryList,
	ipIncludeList *types.IPList,
	ipExcludeList *types.IPList,
	asnIncludeList *types.MapList[uint32],
	asnExcludeList *types.MapList[uint32],
	ja3IncludeList *types.MapList[string],
	ja3ExcludeList *types.MapList[string],
	domainIncludeList *types.DomainList,
	domainExcludeList *types.DomainList,
) []filter.Filter {
	// geofilter.IPLocate is responsible for the ASNList updates
	asnList := types.NewASNList()

	head := []filter.Filter{
		// ip allowlist
		ipfilter.NewAllowList(logger, ipAllowList),
		// quarantine filter
		quarantinefilter.NewQuarantineIP(cfg, natsClient, logger),
		// rate filter
		ratefilter.NewLimiter(cfg.LimiterRate, cfg.LimiterBurst, cfg.LimiterCacheSize, cfg.LimiterBucketTTL, natsClient, logger),
	}

	tail := []filter.Filter{
		// ip filters
		ipfilter.NewFireHOL([]string{
			"https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level1.netset",
		}, logger, ipIncludeList, ipExcludeList),
		ipfilter.NewSpamhaus([]string{
			"https://www.spamhaus.org/drop/drop.txt",
		}, logger, ipIncludeList, ipExcludeList),
		ipfilter.NewAbuseCH([]string{
			"https://feodotracker.abuse.ch/downloads/ipblocklist.txt",
		}, logger, ipIncludeList, ipExcludeList),
		// geo filters
		geofilter.NewIPLocate([]string{
			"https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/ip-to-asn/ip-to-asn.csv.zip",
		}, logger, asnList, countryBlocklist),
		// asn filters
		asnfilter.NewSpamhaus([]string{
			"https://www.spamhaus.org/drop/asndrop.json",
		}, logger, asnList, asnIncludeList, asnExcludeList),
		// dns/sni filters
		domainfilter.NewStevenBlack([]string{
			"https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
		}, natsClient, logger, domainIncludeList, domainExcludeList),
		domainfilter.NewSomeoneWhoCares([]string{
			"https://someonewhocares.org/hosts/hosts",
		}, natsClient, logger, domainIncludeList, domainExcludeList),
		// ja3 filters
		ja3filter.NewAbuseCH([]string{
			"https://sslbl.abuse.ch/blacklist/ja3_fingerprints.csv",
		}, natsClient, logger, ja3IncludeList, ja3ExcludeList),
	}

	if cfg.FilterAbuseIPDBEnable {
		abuseipdb := ipfilter.NewAbuseIPDB([]string{
			"https://api.abuseipdb.com/api/v2/blacklist",
		}, cfg.FilterAbuseIPDBApiKey, cfg.FilterAbuseIPDBConfidence, logger, ipIncludeList, ipExcludeList)

		return append(append(head, abuseipdb), tail...)
	}

	return append(head, tail...)
}

func loadIPLists(ctx context.Context, db *database.Database) (*types.IPList, *types.IPList, error) {
	// includelist
	include, err := db.Q.GetAllIPIncludeList(ctx, db.DB)
	if err != nil {
		return nil, nil, fmt.Errorf("get all includelist: %w", err)
	}
	subnetList, err := get.Subnets(include)
	if err != nil {
		return nil, nil, fmt.Errorf("get subnet includelist: %w", err)
	}
	includeList := types.NewIPList()
	if err := includeList.Upsert(subnetList); err != nil {
		return nil, nil, fmt.Errorf("upsert subnet includelist: %w", err)
	}

	// excludelist
	exclude, err := db.Q.GetAllIPExcludeList(ctx, db.DB)
	if err != nil {
		return nil, nil, fmt.Errorf("get all excludelist: %w", err)
	}
	subnetList, err = get.Subnets(exclude)
	if err != nil {
		return nil, nil, fmt.Errorf("get subnet excludelist: %w", err)
	}
	excludeList := types.NewIPList()
	if err := excludeList.Upsert(subnetList); err != nil {
		return nil, nil, fmt.Errorf("upsert subnet excludelist: %w", err)
	}

	return includeList, excludeList, nil
}

func loadASNLists(ctx context.Context, db *database.Database) (*types.MapList[uint32], *types.MapList[uint32], error) {
	// includelist
	include, err := db.Q.GetAllASNIncludeList(ctx, db.DB)
	if err != nil {
		return nil, nil, fmt.Errorf("get all includelist: %w", err)
	}
	asnList := make([]uint32, 0, len(include))
	for _, asn := range include {
		asnList = append(asnList, uint32(asn))
	}
	includeList := types.NewMapList[uint32]()
	if err := includeList.Upsert(asnList); err != nil {
		return nil, nil, fmt.Errorf("upsert includelist: %w", err)
	}

	// excludelist
	exclude, err := db.Q.GetAllASNExcludeList(ctx, db.DB)
	if err != nil {
		return nil, nil, fmt.Errorf("get all excludelist: %w", err)
	}
	asnList = make([]uint32, 0, len(exclude))
	for _, asn := range exclude {
		asnList = append(asnList, uint32(asn))
	}
	excludeList := types.NewMapList[uint32]()
	if err := excludeList.Upsert(asnList); err != nil {
		return nil, nil, fmt.Errorf("upsert excludelist: %w", err)
	}

	return includeList, excludeList, nil
}

func loadJA3Lists(ctx context.Context, db *database.Database) (*types.MapList[string], *types.MapList[string], error) {
	// includelist
	include, err := db.Q.GetAllJA3IncludeList(ctx, db.DB)
	if err != nil {
		return nil, nil, fmt.Errorf("get all includelist: %w", err)
	}
	includeList := types.NewMapList[string]()
	if err := includeList.Upsert(include); err != nil {
		return nil, nil, fmt.Errorf("upsert includelist: %w", err)
	}

	// excludelist
	exclude, err := db.Q.GetAllJA3ExcludeList(ctx, db.DB)
	if err != nil {
		return nil, nil, fmt.Errorf("get all excludelist: %w", err)
	}
	excludeList := types.NewMapList[string]()
	if err := excludeList.Upsert(exclude); err != nil {
		return nil, nil, fmt.Errorf("upsert excludelist: %w", err)
	}

	return includeList, excludeList, nil
}

func loadDomainLists(ctx context.Context, db *database.Database) (*types.DomainList, *types.DomainList, error) {
	// includelist
	include, err := db.Q.GetAllDomainIncludeList(ctx, db.DB)
	if err != nil {
		return nil, nil, fmt.Errorf("get all includelist: %w", err)
	}
	includeList := types.NewDomainList()
	if err := includeList.Upsert(include); err != nil {
		return nil, nil, fmt.Errorf("upsert includelist: %w", err)
	}

	// excludelist
	exclude, err := db.Q.GetAllDomainExcludeList(ctx, db.DB)
	if err != nil {
		return nil, nil, fmt.Errorf("get all excludelist: %w", err)
	}
	excludeList := types.NewDomainList()
	if err := excludeList.Upsert(exclude); err != nil {
		return nil, nil, fmt.Errorf("upsert excludelist: %w", err)
	}

	return includeList, excludeList, nil
}

func loadAllowBlockLists(ctx context.Context, db *database.Database) (*types.IPList, *types.CountryList, error) {
	// load ip allowlist
	snAllowList, err := db.Q.GetAllIPAllowList(ctx, db.DB)
	if err != nil {
		return nil, nil, fmt.Errorf("get all ip allowlist: %w", err)
	}
	subnets, err := get.Subnets(snAllowList)
	if err != nil {
		return nil, nil, fmt.Errorf("get subnet allowlist: %w", err)
	}
	ipAllowList := types.NewIPList()
	if err := ipAllowList.Upsert(subnets); err != nil {
		return nil, nil, fmt.Errorf("upsert subnet allowlist: %w", err)
	}

	// load country blocklist
	crBlockList, err := db.Q.GetAllCountryBlockList(ctx, db.DB)
	if err != nil {
		return nil, nil, fmt.Errorf("get all country blocklist: %w", err)
	}
	countryBlockList := types.NewCountryList()
	if err := countryBlockList.Upsert(crBlockList); err != nil {
		return nil, nil, fmt.Errorf("upsert country blocklist: %w", err)
	}

	return ipAllowList, countryBlockList, nil
}
