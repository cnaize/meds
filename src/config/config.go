package config

import (
	"time"
)

type Config struct {
	LogLevel   string
	DBFilePath string
	// core
	WorkersCount uint
	LoggersCount uint
	// filter
	UpdateTimeout  time.Duration
	UpdateInterval time.Duration
	// api server
	Username      string
	Password      string
	APIServerAddr string
	// rate limiter
	LimiterMaxBalance uint
	LimiterRefillRate uint
	LimiterCacheSize  uint
	LimiterBucketTTL  time.Duration
}
