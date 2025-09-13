package get

import (
	"net/netip"
	"slices"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func PacketSrcIP(packet gopacket.Packet) (netip.Addr, bool) {
	ip4, ok := packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	if !ok {
		return netip.Addr{}, false
	}

	return netip.AddrFrom4(*(*[4]byte)(ip4.SrcIP.To4())), true
}

func NetPrefix(str string) (netip.Prefix, bool) {
	if str == "" {
		return netip.Prefix{}, false
	}

	if prefix, err := netip.ParsePrefix(str); err == nil {
		return prefix, true
	}

	if ip, err := netip.ParseAddr(str); err == nil && ip.Is4() {
		return netip.PrefixFrom(ip, 32), true
	}

	return netip.Prefix{}, false
}

func ReversedDomain(domain string) string {
	parts := strings.Split(strings.ToLower(domain), ".")
	slices.Reverse(parts)
	return strings.Join(parts, ".")
}
