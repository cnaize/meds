package config

import "time"

type Config struct {
	QueueCount     uint
	UpdateTimeout  time.Duration
	UpdateInterval time.Duration
}
