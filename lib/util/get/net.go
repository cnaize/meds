package get

import (
	"fmt"
	"net/netip"
	"slices"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/cnaize/meds/lib/util"
)

func Proto(packet gopacket.Packet) (layers.IPProtocol, bool) {
	ip4, ok := packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	if !ok {
		return 0, false
	}

	return ip4.Protocol, true
}

func SrcIP(packet gopacket.Packet) (netip.Addr, bool) {
	ip4, ok := packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	if !ok || len(ip4.SrcIP) != 4 {
		return netip.Addr{}, false
	}

	return netip.AddrFrom4(*(*[4]byte)(ip4.SrcIP)), true
}

func DstIP(packet gopacket.Packet) (netip.Addr, bool) {
	ip4, ok := packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	if !ok || len(ip4.DstIP) != 4 {
		return netip.Addr{}, false
	}

	return netip.AddrFrom4(*(*[4]byte)(ip4.DstIP)), true
}

func SrcPort(packet gopacket.Packet) (uint16, bool) {
	tcp, ok := packet.Layer(layers.LayerTypeTCP).(*layers.TCP)
	if ok {
		return uint16(tcp.SrcPort), true
	}

	udp, ok := packet.Layer(layers.LayerTypeUDP).(*layers.UDP)
	if ok {
		return uint16(udp.SrcPort), true
	}

	return 0, false
}

func DstPort(packet gopacket.Packet) (uint16, bool) {
	tcp, ok := packet.Layer(layers.LayerTypeTCP).(*layers.TCP)
	if ok {
		return uint16(tcp.DstPort), true
	}

	udp, ok := packet.Layer(layers.LayerTypeUDP).(*layers.UDP)
	if ok {
		return uint16(udp.DstPort), true
	}

	return 0, false
}

func Subnet(str string) (netip.Prefix, bool) {
	if prefix, err := netip.ParsePrefix(str); err == nil {
		return prefix, true
	}

	if ip, err := netip.ParseAddr(str); err == nil && ip.Is4() {
		return netip.PrefixFrom(ip, 32), true
	}

	return netip.Prefix{}, false
}

func Subnets(strs []string) ([]netip.Prefix, error) {
	subnets := make([]netip.Prefix, len(strs))
	for i, str := range strs {
		subnet, ok := Subnet(str)
		if !ok {
			return nil, fmt.Errorf("parse: %s", str)
		}
		subnets[i] = subnet
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

		questions = append(questions, util.BytesToString(question.Name))
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

		answers = append(answers, util.BytesToString(answer.CNAME))
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

		domains = append(domains, util.BytesToString(question.Name))
	}
	// collect answers
	for _, answer := range dns.Answers {
		if len(answer.CNAME) < 1 {
			continue
		}

		domains = append(domains, util.BytesToString(answer.CNAME))
	}

	return domains
}

func ReversedDomain(domain string) string {
	parts := strings.Split(strings.ToLower(domain), ".")
	slices.Reverse(parts)
	return strings.Join(parts, ".")
}
