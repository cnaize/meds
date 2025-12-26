package rate

import (
	"sync/atomic"
	"time"
)

type Bucket struct {
	balance atomic.Int64
	updated atomic.Int64
}

func NewBucket(burst uint) *Bucket {
	return (&Bucket{}).Reset(burst)
}

func (b *Bucket) Reset(burst uint) *Bucket {
	b.balance.Store(int64(burst))
	b.updated.Store(time.Now().UnixNano())

	return b
}

// approximate but fast
func (b *Bucket) Allow(rate, burst uint) bool {
	now := time.Now().UnixNano()
	updated := b.updated.Load()
	elapsed := now - updated

	add := (elapsed * int64(rate)) / int64(time.Second)
	if add > 0 {
		elapsed = (add * int64(time.Second)) / int64(rate)
		if b.updated.CompareAndSwap(updated, updated+elapsed) {
			if b.balance.Add(add) > int64(burst) {
				b.balance.Store(int64(burst))
			}
		}
	}

	if b.balance.Add(-1) < 0 {
		b.balance.Add(1)
		return false
	}

	return true
}
