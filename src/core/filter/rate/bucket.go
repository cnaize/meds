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
	var bucket Bucket
	bucket.balance.Store(int64(maxBalance))
	bucket.updated.Store(time.Now().UnixNano())

	return &bucket
}

// approximate but fast
func (t *Bucket) Allow(maxBalance, refillRate uint) bool {
	now := time.Now().UnixNano()
	updated := t.updated.Load()
	if now > updated && t.updated.CompareAndSwap(updated, now) {
		elapsed := float64(now-updated) / float64(time.Second)
		add := int64(elapsed * float64(refillRate))
		if add > 0 {
			balance := t.balance.Add(add)
			if balance > int64(maxBalance) {
				t.balance.Store(int64(maxBalance))
			}
		}
	}

	if t.balance.Add(-1) < 0 {
		t.balance.Add(1)
		return false
	}

	return true
}
