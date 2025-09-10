package config

import "time"

type Config struct {
	QCount         uint
	UpdateInterval time.Duration
}
