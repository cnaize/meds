package rate

import (
	"context"
	"fmt"
	"net/netip"
	"sync"
	"time"

	"github.com/maypok86/otter/v2"
	"github.com/nats-io/nats.go"

	"github.com/cnaize/meds/pkg"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/core/metrics"
	"github.com/cnaize/meds/src/types"
)

var _ filter.Filter = (*Limiter)(nil)

type Limiter struct {
	rate      uint
	burst     uint
	cacheSize uint
	bucketTTL time.Duration

	nc     *nats.Conn
	logger *logger.Logger

	cache *otter.Cache[netip.Addr, *Bucket]
	bpool sync.Pool
}

func NewLimiter(rate, burst, cacheSize uint, bucketTTL time.Duration, nc *nats.Conn, logger *logger.Logger) *Limiter {
	return &Limiter{
		rate:      rate,
		burst:     burst,
		cacheSize: cacheSize,
		bucketTTL: bucketTTL,
		nc:        nc,
		logger:    logger,
		bpool: sync.Pool{
			New: func() any {
				return NewBucket(burst)
			},
		},
	}
}

func (f *Limiter) Name() string {
	return "Limiter"
}

func (f *Limiter) Type() filter.FilterType {
	return filter.FilterTypeRate
}

func (f *Limiter) Load(ctx context.Context) error {
	var err error
	f.cache, err = otter.New(
		&otter.Options[netip.Addr, *Bucket]{
			MaximumSize:      int(f.cacheSize),
			ExpiryCalculator: otter.ExpiryAccessing[netip.Addr, *Bucket](f.bucketTTL),
			OnDeletion: func(e otter.DeletionEvent[netip.Addr, *Bucket]) {
				f.bpool.Put(e.Value)
			},
			OnAtomicDeletion: func(e otter.DeletionEvent[netip.Addr, *Bucket]) {
				f.bpool.Put(e.Value)
			},
			StatsRecorder: metrics.Get().RateLimiterCacheStats,
		},
	)
	if err != nil {
		return fmt.Errorf("new cache: %w", err)
	}

	f.logger.Raw().Info().Str("name", f.Name()).Str("type", string(f.Type())).Msg("Filter loaded")

	return nil
}

func (f *Limiter) Check(packet *types.Packet) bool {
	srcIP, ok := packet.GetSrcIP()
	if !ok {
		return true
	}

	bucket, err := f.cache.Get(context.Background(), srcIP,
		otter.LoaderFunc[netip.Addr, *Bucket](
			func(ctx context.Context, key netip.Addr) (*Bucket, error) {
				return f.bpool.Get().(*Bucket).Reset(f.burst), nil
			},
		),
	)
	if err != nil {
		f.logger.Raw().Warn().Err(err).Str("name", f.Name()).Str("type", string(f.Type())).Msg("get bucket failed")
		return true
	}

	if !bucket.Allow(f.rate, f.burst) {
		// add src ip to quarantine
		f.nc.Publish(pkg.NatsSubjectQuarantineIPAdd, []byte(srcIP.String()))
		return false
	}

	return true
}

func (f *Limiter) Update(ctx context.Context) error {
	return nil
}
