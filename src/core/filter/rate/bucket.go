package rate

import (
	"sync/atomic"
	"time"
)

type Bucket struct {
	balance atomic.Int64
	updated atomic.Int64
}

func NewBucket(maxBalance uint) *Bucket {
	return (&Bucket{}).Reset(maxBalance)
}

// approximate but fast
func (b *Bucket) Allow(maxBalance, refillRate uint) bool {
	now := time.Now().UnixNano()
	updated := b.updated.Load()
	if now > updated && b.updated.CompareAndSwap(updated, now) {
		elapsed := float64(now-updated) / float64(time.Second)
		add := int64(elapsed * float64(refillRate))
		if add > 0 {
			balance := b.balance.Add(add)
			if balance > int64(maxBalance) {
				b.balance.Store(int64(maxBalance))
			}
		}
	}

	if b.balance.Add(-1) < 0 {
		b.balance.Add(1)
		return false
	}

	return true
}

func (b *Bucket) Reset(maxBalance uint) *Bucket {
	b.balance.Store(int64(maxBalance))
	b.updated.Store(time.Now().UnixNano())

	return b
}
