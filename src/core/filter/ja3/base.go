package ja3

import (
	"context"
	"sync/atomic"

	"github.com/google/gopacket"

	"github.com/cnaize/meds/lib/util/get"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
)

type Base struct {
	urls      []string
	logger    *logger.Logger
	blacklist atomic.Pointer[map[string]bool]
}

func NewBase(urls []string, logger *logger.Logger) *Base {
	return &Base{
		urls:   urls,
		logger: logger,
	}
}

func (f *Base) Type() filter.FilterType {
	return filter.FilterTypeJA3
}

func (f *Base) Load(ctx context.Context) error {
	f.blacklist.Store(new(map[string]bool))

	return nil
}

func (f *Base) Check(packet gopacket.Packet) bool {
	hash, ok := get.JA3(packet)
	if !ok {
		return true
	}

	list := f.blacklist.Load()
	return !(*list)[hash]
}
