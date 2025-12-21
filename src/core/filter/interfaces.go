package filter

import (
	"context"

	"github.com/cnaize/meds/src/types"
)

type FilterType string

const (
	FilterTypeEmpty  FilterType = "empty"
	FilterTypeIP     FilterType = "ip"
	FilterTypeASN    FilterType = "asn"
	FilterTypeJA3    FilterType = "ja3"
	FilterTypeMeta   FilterType = "meta"
	FilterTypeRate   FilterType = "rate"
	FilterTypeDomain FilterType = "domain"
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
	Check(packet *types.Packet) bool
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
