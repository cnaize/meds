package types

import (
	"net/netip"
	"sync/atomic"

	"github.com/gaissmai/bart"
)

type ASN struct {
	ASN     uint32
	Country string
}

type ASNList struct {
	list atomic.Pointer[bart.Table[ASN]]
}

func NewASNList() *ASNList {
	var l ASNList
	l.list.Store(new(bart.Table[ASN]))

	return &l
}

func (l *ASNList) Lookup(addr netip.Addr) (ASN, bool) {
	return l.list.Load().Lookup(addr)
}

func (l *ASNList) Load() *bart.Table[ASN] {
	return l.list.Load()
}

func (l *ASNList) Store(list *bart.Table[ASN]) {
	l.list.Store(list)
}
