package core

import (
	"context"

	"github.com/google/gopacket"
)

type Loader interface {
	Load(ctx context.Context) error
}

type Checker interface {
	Check(packet gopacket.Packet) bool
}

type Updater interface {
	Update(ctx context.Context) error
}

type Filter interface {
	Loader
	Checker
	Updater
}
