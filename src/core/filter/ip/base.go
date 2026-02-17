package ip

import (
	"context"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/types"
)

type Base struct {
	urls   []string
	logger *logger.Logger

	include   *types.IPList
	exclude   *types.IPList
	blocklist *types.IPList
}

func NewBase(urls []string, logger *logger.Logger, include, exclude *types.IPList) *Base {
	return &Base{
		urls:    urls,
		logger:  logger,
		include: include,
		exclude: exclude,
	}
}

func (f *Base) Type() filter.FilterType {
	return filter.FilterTypeIP
}

func (f *Base) Load(ctx context.Context) error {
	f.blocklist = types.NewIPList()

	return nil
}

func (f *Base) Check(packet *types.Packet) bool {
	srcIP, ok := packet.GetSrcIP()
	if !ok {
		return true
	}

	// check excludelist
	if f.exclude.Lookup(srcIP) {
		return true
	}

	// check include/block lists
	return !(f.include.Lookup(srcIP) || f.blocklist.Lookup(srcIP))
}
