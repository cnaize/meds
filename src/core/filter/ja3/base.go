package ja3

import (
	"context"
	"sync/atomic"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/types"
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

func (f *Base) Check(packet *types.Packet) bool {
	hash, ok := packet.GetJA3()
	if !ok {
		return true
	}

	list := f.blacklist.Load()
	return !(*list)[hash]
}
