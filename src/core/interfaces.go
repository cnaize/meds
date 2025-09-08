package core

import (
	"context"

	"github.com/google/gopacket"
)

type Checker interface {
	Check(packet gopacket.Packet) bool
}

type Updater interface {
	Update(ctx context.Context) error
}

type Filter interface {
	Checker
	Updater
}
