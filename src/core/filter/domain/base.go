package domain

import (
	"context"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/types"
)

type Base struct {
	urls   []string
	logger *logger.Logger

	include   *types.DomainList
	exclude   *types.DomainList
	blocklist *types.DomainList
}

func NewBase(urls []string, logger *logger.Logger, include, exclude *types.DomainList) *Base {
	return &Base{
		urls:    urls,
		logger:  logger,
		include: include,
		exclude: exclude,
	}
}

func (f *Base) Type() filter.FilterType {
	return filter.FilterTypeDomain
}

func (f *Base) Load(ctx context.Context) error {
	f.blocklist = types.NewDomainList()

	return nil
}

func (f *Base) Check(packet *types.Packet) bool {
	for _, domain := range packet.GetDomains() {
		// check excludelist
		if f.exclude.Lookup(domain) {
			continue
		}

		// check include/block lists
		if f.include.Lookup(domain) || f.blocklist.Lookup(domain) {
			return false
		}
	}

	return true
}
