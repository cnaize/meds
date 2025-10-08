package get

import (
	"net/netip"
	"slices"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/cnaize/meds/lib/util"
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

func DNSItems(packet gopacket.Packet) []string {
	dns, ok := packet.Layer(layers.LayerTypeDNS).(*layers.DNS)
	if !ok {
		return nil
	}

	items := make([]string, 0, len(dns.Questions)+len(dns.Answers))
	// collect questions
	for _, question := range dns.Questions {
		if len(question.Name) < 1 {
			continue
		}

		items = append(items, util.BytesToString(question.Name))
	}
	// collect answers
	for _, answer := range dns.Answers {
		if len(answer.CNAME) < 1 {
			continue
		}

		items = append(items, util.BytesToString(answer.CNAME))
	}

	return items
}

func ReversedDomain(domain string) string {
	parts := strings.Split(strings.ToLower(domain), ".")
	slices.Reverse(parts)
	return strings.Join(parts, ".")
}
