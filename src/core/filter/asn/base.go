package asn

import (
	"context"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/types"
)

type Base struct {
	urls   []string
	logger *logger.Logger

	asnlist *types.ASNList

	include   *types.MapList[uint32]
	exclude   *types.MapList[uint32]
	blocklist *types.MapList[uint32]
}

func NewBase(urls []string, logger *logger.Logger, asnlist *types.ASNList, include, exclude *types.MapList[uint32]) *Base {
	return &Base{
		urls:    urls,
		logger:  logger,
		asnlist: asnlist,
		include: include,
		exclude: exclude,
	}
}

func (f *Base) Type() filter.FilterType {
	return filter.FilterTypeASN
}

func (f *Base) Load(ctx context.Context) error {
	f.blocklist = types.NewMapList[uint32]()

	return nil
}

func (f *Base) Check(packet *types.Packet) bool {
	asn, ok := packet.GetASN(f.asnlist)
	if !ok {
		return true
	}

	// check excludelist
	if f.exclude.Lookup(asn.ASN) {
		return true
	}

	// check include/block lists
	return !(f.include.Lookup(asn.ASN) || f.blocklist.Lookup(asn.ASN))
}
