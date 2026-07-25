package quarantine

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/maypok86/otter/v2"
	"github.com/nats-io/nats.go"

	"github.com/cnaize/meds/pkg"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/core/metrics"
	"github.com/cnaize/meds/src/types"
)

var _ filter.Filter = (*QuarantineIP)(nil)

type QuarantineIP struct {
	cacheSize uint
	entityTTL time.Duration

	nc     *nats.Conn
	logger *logger.Logger

	sub   *nats.Subscription
	cache *otter.Cache[netip.Addr, struct{}]
}

func NewQuarantineIP(cacheSize uint, entityTTL time.Duration, nc *nats.Conn, logger *logger.Logger) *QuarantineIP {
	return &QuarantineIP{
		cacheSize: cacheSize,
		entityTTL: entityTTL,
		nc:        nc,
		logger:    logger,
	}
}

func (f *QuarantineIP) Name() string {
	return "Quarantine"
}

func (f *QuarantineIP) Type() filter.FilterType {
	return filter.FilterTypeIP
}

func (f *QuarantineIP) Load(ctx context.Context) error {
	var err error
	f.cache, err = otter.New(
		&otter.Options[netip.Addr, struct{}]{
			MaximumSize:      int(f.cacheSize),
			ExpiryCalculator: otter.ExpiryAccessing[netip.Addr, struct{}](f.entityTTL),
			StatsRecorder:    metrics.Get().QuarantineIPCacheStats,
		},
	)
	if err != nil {
		return fmt.Errorf("new cache: %w", err)
	}

	f.sub, err = f.nc.SubscribeSync(pkg.NatsSubjectQuarantineIP + ".*")
	if err != nil {
		return fmt.Errorf("%s: subscribe subject: %w", pkg.NatsSubjectQuarantineIP, err)
	}

	go f.natsHandler(ctx)

	f.logger.Raw().Info().Str("name", f.Name()).Str("type", string(f.Type())).Msg("Filter loaded")

	return nil
}

func (f *QuarantineIP) Check(packet *types.Packet) bool {
	srcIP, ok := packet.GetSrcIP()
	if !ok {
		return true
	}

	if _, found := f.cache.GetIfPresent(srcIP); found {
		return false
	}

	return true
}

func (f *QuarantineIP) Update(ctx context.Context) error {
	return nil
}

func (f *QuarantineIP) natsHandler(ctx context.Context) {
	defer f.sub.Unsubscribe()

	for {
		msg, err := f.sub.NextMsgWithContext(ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				f.logger.Raw().Warn().Err(err).Str("name", f.Name()).Str("type", string(f.Type())).Msg("get nats msg failed")
			}
			return
		}

		ip, err := netip.ParseAddr(string(msg.Data))
		if err != nil {
			f.logger.Raw().Warn().Err(err).Str("name", f.Name()).Str("type", string(f.Type())).Msg("parse nats msg failed")
			continue
		}

		if !ip.Is4() {
			continue
		}

		switch msg.Subject {
		case pkg.NatsSubjectQuarantineIPAdd:
			f.cache.Set(ip, struct{}{})
		case pkg.NatsSubjectQuarantineIPDel:
			f.cache.Invalidate(ip)
		}
	}
}
