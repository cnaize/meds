package config

import (
	"time"
)

type Config struct {
	LogLevel       string
	MetricsAddr    string
	EnableMetrics  bool
	WorkersCount   uint
	LoggersCount   uint
	UpdateTimeout  time.Duration
	UpdateInterval time.Duration
}
