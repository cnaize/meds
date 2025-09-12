package ip

import (
	"context"
	"net/netip"
	"sync/atomic"

	"github.com/gaissmai/bart"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
)

type Base struct {
	urls      []string
	logger    *logger.Logger
	blackList atomic.Pointer[bart.Lite]
}

func NewBase(urls []string, logger *logger.Logger) *Base {
	return &Base{
		urls:      urls,
		logger:    logger,
		blackList: atomic.Pointer[bart.Lite]{},
	}
}

func (f *Base) Type() filter.FilterType {
	return filter.FilterTypeIP
}

func (f *Base) Load(ctx context.Context) error {
	f.blackList.Store(new(bart.Lite))

	return nil
}

func (f *Base) Check(packet gopacket.Packet) bool {
	ip4, ok := packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	if !ok {
		return true
	}

	list := f.blackList.Load()
	srcIP := netip.AddrFrom4(*(*[4]byte)(ip4.SrcIP.To4()))
	if list.Contains(srcIP) {
		return false
	}

	return true
}

func ParsePrefix(str string) (netip.Prefix, bool) {
	prefix, err := netip.ParsePrefix(str)
	if err != nil {
		ip, err := netip.ParseAddr(str)
		if err != nil {
			return netip.Prefix{}, false
		}

		if !ip.Is4() {
			return netip.Prefix{}, false
		}

		prefix = netip.PrefixFrom(ip, 32)
	}

	return prefix, true
}
