package domain

import (
	"context"
	"sync/atomic"

	"github.com/armon/go-radix"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/types"
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

func (f *Base) Check(packet *types.Packet) bool {
	list := f.blacklist.Load()
	for _, revDomain := range packet.GetReversedDomains() {
		if _, _, found := list.LongestPrefix(revDomain); found {
			return false
		}
	}

	return true
}
