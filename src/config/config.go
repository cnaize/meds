package config

import (
	"time"
)

type Config struct {
	LogLevel       string
	MetricsAddr    string
	WorkersCount   uint
	LoggersCount   uint
	UpdateTimeout  time.Duration
	UpdateInterval time.Duration
	// rate limiter
	LimiterMaxBalance uint
	LimiterRefillRate uint
	LimiterCacheSize  uint
	LimiterBucketTTL  time.Duration
}
