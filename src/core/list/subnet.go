package list

import (
	"fmt"
	"sync/atomic"

	"github.com/gaissmai/bart"

	"github.com/cnaize/meds/lib/util/get"
)

var _ List = (*SubnetList)(nil)

// WARNING: lock free, but NOT thread safe
type SubnetList struct {
	list atomic.Pointer[bart.Lite]
}

func NewSubnetList() *SubnetList {
	var l SubnetList
	l.list.Store(new(bart.Lite))

	return &l
}

func (l *SubnetList) Lookup(subnet string) bool {
	prefix, ok := get.NetPrefix(subnet)
	if !ok {
		return false
	}

	return l.list.Load().Exists(prefix)
}

func (l *SubnetList) Upsert(subnets []string) error {
	// clone list
	list := l.list.Load().Clone()
	// insert provided items
	for _, subnet := range subnets {
		prefix, ok := get.NetPrefix(subnet)
		if !ok {
			return fmt.Errorf("invalid subnet: %s", subnet)
		}

		list.Insert(prefix)
	}

	l.list.Store(list)

	return nil
}

func (l *SubnetList) Remove(subnets []string) error {
	// clone list
	list := l.list.Load().Clone()
	// delete provided items
	for _, subnet := range subnets {
		prefix, ok := get.NetPrefix(subnet)
		if !ok {
			return fmt.Errorf("invalid subnet: %s", subnet)
		}

		list.Delete(prefix)
	}

	l.list.Store(list)

	return nil
}
