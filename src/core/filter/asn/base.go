package asn

import (
	"context"
	"sync/atomic"

	"github.com/cnaize/meds/lib/util/get"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/types"
)

type Base struct {
	urls   []string
	logger *logger.Logger

	ipToASN   *IPLocate
	blacklist atomic.Pointer[map[uint32]bool]
}

func NewBase(urls []string, logger *logger.Logger, ipToASN *IPLocate) *Base {
	return &Base{
		urls:    urls,
		logger:  logger,
		ipToASN: ipToASN,
	}
}

func (f *Base) Type() filter.FilterType {
	return filter.FilterTypeASN
}

func (f *Base) Load(ctx context.Context) error {
	f.blacklist.Store(get.Ptr(make(map[uint32]bool)))

	return nil
}

func (f *Base) Check(packet *types.Packet) bool {
	// set asn
	packet.SetASN(f.ipToASN.ipToASN.Load())

	// get asn
	asn, ok := packet.GetASN()
	if !ok {
		return true
	}

	// check asn
	list := *f.blacklist.Load()
	return !list[asn]
}
