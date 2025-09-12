package filter

import (
	"context"

	"github.com/google/gopacket"
)

type FilterType string

const (
	FilterTypeIP  FilterType = "ip"
	FilterTypeDNS FilterType = "dns"
)

type Namer interface {
	Name() string
}

type Typer interface {
	Type() FilterType
}

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
	Namer
	Typer

	Loader
	Checker
	Updater
}
