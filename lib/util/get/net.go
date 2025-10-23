package get

import (
	"fmt"
	"net/netip"
	"slices"
	"strings"

	"darvaza.org/x/tls/sni"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/open-ch/ja3"

	"github.com/cnaize/meds/lib/util"
)

func PacketSrcIP(packet gopacket.Packet) (netip.Addr, bool) {
	ip4, ok := packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	if !ok {
		return netip.Addr{}, false
	}

	return netip.AddrFrom4(*(*[4]byte)(ip4.SrcIP.To4())), true
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

func SNI(packet gopacket.Packet) (string, bool) {
	tcp, ok := packet.Layer(layers.LayerTypeTCP).(*layers.TCP)
	if !ok {
		return "", false
	}

	info := sni.GetInfo(tcp.LayerPayload())
	if info == nil {
		return "", false
	}

	if len(info.ServerName) < 1 {
		return "", false
	}

	return info.ServerName, true
}

func Domains(packet gopacket.Packet) []string {
	domains := DNSDomains(packet)
	if sni, ok := SNI(packet); ok {
		domains = append(domains, sni)
	}

	return domains
}

func JA3(packet gopacket.Packet) (string, bool) {
	tcp, ok := packet.Layer(layers.LayerTypeTCP).(*layers.TCP)
	if !ok {
		return "", false
	}

	j, err := ja3.ComputeJA3FromSegment(tcp.LayerPayload())
	if err != nil {
		return "", false
	}

	return j.GetJA3Hash(), true
}

func ReversedDomain(domain string) string {
	parts := strings.Split(strings.ToLower(domain), ".")
	slices.Reverse(parts)
	return strings.Join(parts, ".")
}
