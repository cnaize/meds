package geo

import (
	"context"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/types"
)

type Base struct {
	urls   []string
	logger *logger.Logger

	asnlist   *types.ASNList
	blacklist *types.CountryList
}

func NewBase(urls []string, logger *logger.Logger, asnlist *types.ASNList, blacklist *types.CountryList) *Base {
	return &Base{
		urls:      urls,
		logger:    logger,
		asnlist:   asnlist,
		blacklist: blacklist,
	}
}

func (f *Base) Type() filter.FilterType {
	return filter.FilterTypeGeo
}

func (f *Base) Load(ctx context.Context) error {
	return nil
}

func (f *Base) Check(packet *types.Packet) bool {
	asn, ok := packet.GetASN(f.asnlist)
	if !ok {
		return true
	}

	return !f.blacklist.Lookup(asn.Country)
}
