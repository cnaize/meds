package list

import (
	"sync/atomic"

	"github.com/armon/go-radix"

	"github.com/cnaize/meds/lib/util/get"
)

var _ List = (*DomainList)(nil)

// WARNING: lock free, but NOT thread safe
type DomainList struct {
	list atomic.Pointer[radix.Tree]
}

func NewDomainList() *DomainList {
	var l DomainList
	l.list.Store(radix.New())

	return &l
}

func (l *DomainList) Lookup(domain string) bool {
	domain = get.ReversedDomain(domain)
	if _, _, found := l.list.Load().LongestPrefix(domain); found {
		return true
	}

	return false
}

func (l *DomainList) Upsert(domains []string) error {
	list := radix.New()
	// clone list
	for item := range l.list.Load().ToMap() {
		list.Insert(item, struct{}{})
	}
	// insert provided items
	for _, domain := range domains {
		list.Insert(get.ReversedDomain(domain), struct{}{})
	}

	l.list.Store(list)

	return nil
}

func (l *DomainList) Remove(domains []string) error {
	list := radix.New()
	// clone list
	for item := range l.list.Load().ToMap() {
		list.Insert(item, struct{}{})
	}
	// delete provided items
	for _, domain := range domains {
		list.DeletePrefix(get.ReversedDomain(domain))
	}

	l.list.Store(list)

	return nil
}
