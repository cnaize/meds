package ja3

import (
	"context"

	"github.com/nats-io/nats.go"

	"github.com/cnaize/meds/pkg"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/types"
)

type Base struct {
	urls []string

	nc     *nats.Conn
	logger *logger.Logger

	include   *types.MapList[string]
	exclude   *types.MapList[string]
	blocklist *types.MapList[string]
}

func NewBase(urls []string, nc *nats.Conn, logger *logger.Logger, include, exclude *types.MapList[string]) *Base {
	return &Base{
		urls:    urls,
		nc:      nc,
		logger:  logger,
		include: include,
		exclude: exclude,
	}
}

func (f *Base) Type() filter.FilterType {
	return filter.FilterTypeJA3
}

func (f *Base) Load(ctx context.Context) error {
	f.blocklist = types.NewMapList[string]()

	return nil
}

func (f *Base) Check(packet *types.Packet) bool {
	hash, ok := packet.GetJA3()
	if !ok {
		return true
	}

	// check excludelist
	if f.exclude.Lookup(hash) {
		return true
	}

	// check include/block lists
	if f.include.Lookup(hash) || f.blocklist.Lookup(hash) {
		// add src ip to quarantine
		if srcIP, ok := packet.GetSrcIP(); ok {
			f.nc.Publish(pkg.NatsSubjectQuarantineIPAdd, srcIP.AppendTo(make([]byte, 0, 16)))
		}

		return false
	}

	return true
}
