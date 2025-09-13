package ip

import (
	"context"
	"sync/atomic"

	"github.com/gaissmai/bart"
	"github.com/google/gopacket"

	"github.com/cnaize/meds/lib/util/get"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
)

type Base struct {
	urls      []string
	logger    *logger.Logger
	blacklist atomic.Pointer[bart.Lite]
}

func NewBase(urls []string, logger *logger.Logger) *Base {
	return &Base{
		urls:      urls,
		logger:    logger,
		blacklist: atomic.Pointer[bart.Lite]{},
	}
}

func (f *Base) Type() filter.FilterType {
	return filter.FilterTypeIP
}

func (f *Base) Load(ctx context.Context) error {
	f.blacklist.Store(new(bart.Lite))

	return nil
}

func (f *Base) Check(packet gopacket.Packet) bool {
	srcIP, ok := get.PacketSrcIP(packet)
	if !ok {
		return true
	}

	if f.blacklist.Load().Contains(srcIP) {
		return false
	}

	return true
}
