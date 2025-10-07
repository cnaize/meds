package config

import (
	"time"
)

type Config struct {
	LogLevel       string
	WorkersCount   uint
	LoggersCount   uint
	UpdateTimeout  time.Duration
	UpdateInterval time.Duration
	// api server
	Username      string
	Password      string
	ApiServerAddr string
	// rate limiter
	LimiterMaxBalance uint
	LimiterRefillRate uint
	LimiterCacheSize  uint
	LimiterBucketTTL  time.Duration
}
