package types

import (
	"net/netip"
	"sync/atomic"

	"github.com/gaissmai/bart"
)

type SubnetList struct {
	list atomic.Pointer[bart.Lite]
}

func NewSubnetList() *SubnetList {
	var l SubnetList
	l.Store(new(bart.Lite))

	return &l
}

func (l *SubnetList) Load() *bart.Lite {
	return l.list.Load()
}

func (l *SubnetList) Store(list *bart.Lite) {
	l.list.Store(list)
}

func (l *SubnetList) GetAll() []netip.Prefix {
	list := l.Load()
	subnets := make([]netip.Prefix, 0, list.Size4())
	for subnet := range list.All4() {
		subnets = append(subnets, subnet)
	}

	return subnets
}

func (l *SubnetList) Lookup(subnet netip.Prefix) bool {
	return l.Load().OverlapsPrefix(subnet)
}

func (l *SubnetList) Upsert(subnets []netip.Prefix) error {
	list := l.Load().Clone()
	for _, subnet := range subnets {
		list.Insert(subnet)
	}

	l.Store(list)

	return nil
}

func (l *SubnetList) Remove(subnets []netip.Prefix) error {
	list := l.Load().Clone()
	for _, subnet := range subnets {
		list.Delete(subnet)
	}

	l.Store(list)

	return nil
}
