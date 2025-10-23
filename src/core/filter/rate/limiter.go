package rate

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"github.com/google/gopacket"
	"github.com/maypok86/otter/v2"

	"github.com/cnaize/meds/lib/util/get"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/core/metrics"
)

var _ filter.Filter = (*Limiter)(nil)

type Limiter struct {
	maxBalance uint
	refillRate uint
	cacheSize  uint
	bucketTTL  time.Duration
	logger     *logger.Logger
	cache      *otter.Cache[netip.Addr, *Bucket]
}

func NewLimiter(maxBalance, refillRate, cacheSize uint, bucketTTL time.Duration, logger *logger.Logger) *Limiter {
	return &Limiter{
		maxBalance: maxBalance,
		refillRate: refillRate,
		cacheSize:  cacheSize,
		bucketTTL:  bucketTTL,
		logger:     logger,
	}
}

func (f *Limiter) Name() string {
	return "Limiter"
}

func (f *Limiter) Type() filter.FilterType {
	return filter.FilterTypeRate
}

func (f *Limiter) Load(ctx context.Context) error {
	cache, err := otter.New(
		&otter.Options[netip.Addr, *Bucket]{
			MaximumSize:       int(f.cacheSize),
			ExpiryCalculator:  otter.ExpiryAccessing[netip.Addr, *Bucket](f.bucketTTL),
			RefreshCalculator: otter.RefreshWriting[netip.Addr, *Bucket](time.Second),
			StatsRecorder:     metrics.Get().RateLimiterCacheStats,
		},
	)
	if err != nil {
		return fmt.Errorf("new cache: %w", err)
	}

	f.logger.Raw().Info().Str("name", f.Name()).Str("type", string(f.Type())).Msg("Filter loaded")
	f.cache = cache

	return nil
}

func (f *Limiter) Check(packet gopacket.Packet) bool {
	srcIP, ok := get.PacketSrcIP(packet)
	if !ok {
		return true
	}

	bucket, err := f.cache.Get(context.Background(), srcIP,
		otter.LoaderFunc[netip.Addr, *Bucket](
			func(ctx context.Context, key netip.Addr) (*Bucket, error) {
				return NewBucket(f.maxBalance), nil
			},
		),
	)
	if err != nil {
		f.logger.Raw().Warn().Err(err).Msg("get bucket failed")
		return true
	}

	return bucket.Allow(f.maxBalance, f.refillRate)
}

func (f *Limiter) Update(ctx context.Context) error {
	return nil
}
