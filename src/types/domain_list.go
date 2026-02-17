package types

import (
	"sync/atomic"

	"github.com/armon/go-radix"

	"github.com/cnaize/meds/lib/util/get"
)

type DomainList struct {
	list atomic.Pointer[radix.Tree]
}

func NewDomainList() *DomainList {
	var l DomainList
	l.Store(radix.New())

	return &l
}

func (l *DomainList) Load() *radix.Tree {
	return l.list.Load()
}

func (l *DomainList) Store(list *radix.Tree) {
	l.list.Store(list)
}

func (l *DomainList) GetAll() []string {
	all := l.Load().ToMap()
	domains := make([]string, 0, len(all))
	for domain := range all {
		domains = append(domains, get.ReversedDomain(domain))
	}

	return domains
}

func (l *DomainList) Lookup(domain string) bool {
	domain = get.ReversedDomain(domain)
	if _, _, found := l.Load().LongestPrefix(domain); found {
		return true
	}

	return false
}

func (l *DomainList) Upsert(domains []string) error {
	list := radix.New()
	for item := range l.Load().ToMap() {
		list.Insert(item, struct{}{})
	}

	for _, domain := range domains {
		list.Insert(get.ReversedDomain(domain), struct{}{})
	}

	l.Store(list)

	return nil
}

func (l *DomainList) Remove(domains []string) error {
	list := radix.New()
	for item := range l.Load().ToMap() {
		list.Insert(item, struct{}{})
	}

	for _, domain := range domains {
		list.DeletePrefix(get.ReversedDomain(domain))
	}

	l.Store(list)

	return nil
}
