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
	// filters
	UpdateTimeout  time.Duration
	UpdateInterval time.Duration
	// geo
	GeoBlackList string
	// api server
	Username      string
	Password      string
	APIServerAddr string
	// rate limiter
	LimiterRate      uint
	LimiterBurst     uint
	LimiterCacheSize uint
	LimiterBucketTTL time.Duration
}
