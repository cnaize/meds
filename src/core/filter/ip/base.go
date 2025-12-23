package ip

import (
	"context"
	"sync/atomic"

	"github.com/gaissmai/bart"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/types"
)

type Base struct {
	urls      []string
	logger    *logger.Logger
	blacklist atomic.Pointer[bart.Lite]
}

func NewBase(urls []string, logger *logger.Logger) *Base {
	return &Base{
		urls:   urls,
		logger: logger,
	}
}

func (f *Base) Type() filter.FilterType {
	return filter.FilterTypeIP
}

func (f *Base) Load(ctx context.Context) error {
	f.blacklist.Store(new(bart.Lite))

	return nil
}

func (f *Base) Check(packet *types.Packet) bool {
	srcIP, ok := packet.GetSrcIP()
	if !ok {
		return true
	}

	list := f.blacklist.Load()
	return !list.Contains(srcIP)
}
