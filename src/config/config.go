package config

import (
	"time"
)

type Config struct {
	LogLevel   string
	DBFilePath string
	// core
	ReadersCount uint
	WorkersCount uint
	LoggersCount uint
	ReaderQLen   uint
	LoggerQLen   uint
	AcceptOnFail bool
	// filters
	UpdateTimeout  time.Duration
	UpdateInterval time.Duration
	// api server
	Username      string
	Password      string
	APIServerAddr string
	// nats server
	NatsEnable   bool
	NatsHost     string
	NatsPort     int
	NatsUsername string
	NatsPassword string
	// quarantine filter
	QuarantineIPCacheSize uint
	QuarantineIPEntityTTL time.Duration
	// rate limiter
	LimiterRate      uint
	LimiterBurst     uint
	LimiterCacheSize uint
	LimiterBucketTTL time.Duration
	// abuseipdb filter
	FilterAbuseIPDBEnable     bool
	FilterAbuseIPDBApiKey     string
	FilterAbuseIPDBConfidence int
	FilterAbuseIPDBReportAddr bool
}
