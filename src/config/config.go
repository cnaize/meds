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
	// api server
	Username      string
	Password      string
	APIServerAddr string
	// nats server
	NatsEnabled  bool
	NatsHost     string
	NatsPort     int
	NatsUsername string
	NatsPassword string
	// nats filter
	NatsBlockIPCacheSize uint
	NatsBlockIPEntiryTTL time.Duration
	// rate limiter
	LimiterRate      uint
	LimiterBurst     uint
	LimiterCacheSize uint
	LimiterBucketTTL time.Duration
}
