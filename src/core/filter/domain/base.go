package domain

import (
	"context"
	"sync/atomic"

	"github.com/armon/go-radix"
	"github.com/google/gopacket"

	"github.com/cnaize/meds/lib/util/get"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
)

type Base struct {
	urls      []string
	logger    *logger.Logger
	blacklist atomic.Pointer[radix.Tree]
}

func NewBase(urls []string, logger *logger.Logger) *Base {
	return &Base{
		urls:   urls,
		logger: logger,
	}
}

func (f *Base) Type() filter.FilterType {
	return filter.FilterTypeDomain
}

func (f *Base) Load(ctx context.Context) error {
	f.blacklist.Store(radix.New())

	return nil
}

func (f *Base) Check(packet gopacket.Packet) bool {
	list := f.blacklist.Load()
	for _, domain := range get.Domains(packet) {
		domain = get.ReversedDomain(domain)
		if _, _, found := list.LongestPrefix(domain); found {
			return false
		}
	}

	return true
}
