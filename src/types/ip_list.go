package types

import (
	"net/netip"
)

type IPList struct {
	*SubnetList
}

func NewIPList() *IPList {
	return &IPList{
		SubnetList: NewSubnetList(),
	}
}

func (l *IPList) Lookup(ip netip.Addr) bool {
	return l.Load().Contains(ip)
}
