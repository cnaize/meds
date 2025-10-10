package types

import (
	"net/netip"
	"sync/atomic"

	"github.com/gaissmai/bart"
)

// WARNING: lock free, but NOT thread safe
type SubnetList struct {
	list atomic.Pointer[bart.Lite]
}

func NewSubnetList() *SubnetList {
	var l SubnetList
	l.list.Store(new(bart.Lite))

	return &l
}

func (l *SubnetList) GetAll() []netip.Prefix {
	list := l.list.Load()
	subnets := make([]netip.Prefix, 0, list.Size4())
	for subnet := range list.All4() {
		subnets = append(subnets, subnet)
	}

	return subnets
}

func (l *SubnetList) Lookup(subnet netip.Prefix) bool {
	return l.list.Load().OverlapsPrefix(subnet)
}

func (l *SubnetList) Upsert(subnets []netip.Prefix) error {
	list := l.list.Load().Clone()
	for _, subnet := range subnets {
		list.Insert(subnet)
	}

	l.list.Store(list)

	return nil
}

func (l *SubnetList) Remove(subnets []netip.Prefix) error {
	list := l.list.Load().Clone()
	for _, subnet := range subnets {
		list.Delete(subnet)
	}

	l.list.Store(list)

	return nil
}
