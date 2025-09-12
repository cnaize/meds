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
}
