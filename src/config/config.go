package config

import "time"

type Config struct {
	WorkersCount   uint
	LoggersCount   uint
	UpdateTimeout  time.Duration
	UpdateInterval time.Duration
}
