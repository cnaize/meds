package get

import (
	"fmt"
	"net/netip"
	"slices"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func Proto(packet gopacket.Packet) (layers.IPProtocol, bool) {
	if ip4, ok := packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4); ok {
		return ip4.Protocol, true
	}

	return 0, false
}

func SrcIP(packet gopacket.Packet) (netip.Addr, bool) {
	ip4, ok := packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	if !ok {
		return netip.Addr{}, false
	}

	ip, ok := netip.AddrFromSlice(ip4.SrcIP)
	if !ok {
		return netip.Addr{}, false
	}

	return ip.Unmap(), true
}

func DstIP(packet gopacket.Packet) (netip.Addr, bool) {
	ip4, ok := packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	if !ok {
		return netip.Addr{}, false
	}

	ip, ok := netip.AddrFromSlice(ip4.DstIP)
	if !ok {
		return netip.Addr{}, false
	}

	return ip.Unmap(), true
}

func SrcPort(packet gopacket.Packet) (uint16, bool) {
	if tcp, ok := packet.Layer(layers.LayerTypeTCP).(*layers.TCP); ok {
		return uint16(tcp.SrcPort), true
	}

	if udp, ok := packet.Layer(layers.LayerTypeUDP).(*layers.UDP); ok {
		return uint16(udp.SrcPort), true
	}

	return 0, false
}

func DstPort(packet gopacket.Packet) (uint16, bool) {
	if tcp, ok := packet.Layer(layers.LayerTypeTCP).(*layers.TCP); ok {
		return uint16(tcp.DstPort), true
	}

	if udp, ok := packet.Layer(layers.LayerTypeUDP).(*layers.UDP); ok {
		return uint16(udp.DstPort), true
	}

	return 0, false
}

func Subnet(str string) (netip.Prefix, bool) {
	if prefix, err := netip.ParsePrefix(str); err == nil {
		return prefix, true
	}

	ip, err := netip.ParseAddr(str)
	if err != nil {
		return netip.Prefix{}, false
	}
	ip = ip.Unmap()

	if !ip.Is4() {
		return netip.Prefix{}, false
	}

	return netip.PrefixFrom(ip, 32), true
}

func Subnets(strs []string) ([]netip.Prefix, error) {
	subnets := make([]netip.Prefix, 0, len(strs))
	for _, str := range strs {
		subnet, ok := Subnet(str)
		if !ok {
			return nil, fmt.Errorf("parse: %s", str)
		}
		subnets = append(subnets, subnet)
	}

	return subnets, nil
}

func DNSQuestions(packet gopacket.Packet) []string {
	dns, ok := packet.Layer(layers.LayerTypeDNS).(*layers.DNS)
	if !ok {
		return nil
	}

	questions := make([]string, 0, len(dns.Questions))
	for _, question := range dns.Questions {
		if len(question.Name) < 1 {
			continue
		}

		questions = append(questions, string(question.Name))
	}

	return questions
}

func DNSAnswers(packet gopacket.Packet) []string {
	dns, ok := packet.Layer(layers.LayerTypeDNS).(*layers.DNS)
	if !ok {
		return nil
	}

	answers := make([]string, 0, len(dns.Answers))
	for _, answer := range dns.Answers {
		if len(answer.CNAME) < 1 {
			continue
		}

		answers = append(answers, string(answer.CNAME))
	}

	return answers
}

func DNSDomains(packet gopacket.Packet) []string {
	dns, ok := packet.Layer(layers.LayerTypeDNS).(*layers.DNS)
	if !ok {
		return nil
	}

	domains := make([]string, 0, len(dns.Questions)+len(dns.Answers))
	// collect questions
	for _, question := range dns.Questions {
		if len(question.Name) < 1 {
			continue
		}

		domains = append(domains, string(question.Name))
	}
	// collect answers
	for _, answer := range dns.Answers {
		if len(answer.CNAME) < 1 {
			continue
		}

		domains = append(domains, string(answer.CNAME))
	}

	return domains
}

func ReversedDomain(domain string) string {
	parts := strings.Split(strings.ToLower(domain), ".")
	slices.Reverse(parts)
	return strings.Join(parts, ".")
}
