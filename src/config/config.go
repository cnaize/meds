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
	// filter
	UpdateTimeout  time.Duration
	UpdateInterval time.Duration
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
