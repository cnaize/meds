package types

import (
	"net/netip"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/open-ch/ja3"

	"github.com/cnaize/meds/lib/util/get"
)

type Packet struct {
	packet     gopacket.Packet
	ja3        *ja3.JA3
	domains    []string
	revDomains []string
}

func NewPacket(payload []byte) (*Packet, error) {
	// WARNING:
	// 1. DON'T MODIFY PACKET (NoCopy: true)
	// 2. NOT THREAD SAFE (Lazy: true)
	packet := gopacket.NewPacket(payload, layers.LayerTypeIPv4, gopacket.DecodeOptions{NoCopy: true, Lazy: true})
	if err := packet.ErrorLayer(); err != nil {
		return nil, err.Error()
	}

	return &Packet{
		packet: packet,
	}, nil
}

func (p *Packet) GetSrcIP() (netip.Addr, bool) {
	return get.SrcIP(p.packet)
}

func (p *Packet) GetDomains() []string {
	// get from cache
	if p.domains != nil {
		return p.domains
	}

	// collect domains
	domains := get.DNSDomains(p.packet)
	if sni, ok := p.GetSNI(); ok && len(sni) > 0 {
		domains = append(domains, sni)
	}

	// save to cache
	p.domains = domains

	return p.domains
}

func (p *Packet) GetReversedDomains() []string {
	// get from cache
	if p.revDomains != nil {
		return p.revDomains
	}

	// reverse domains
	domains := p.GetDomains()
	revDomains := make([]string, len(domains))
	for i, domain := range domains {
		revDomains[i] = get.ReversedDomain(domain)
	}

	// save to cache
	p.revDomains = revDomains

	return p.revDomains
}

func (p *Packet) GetSNI() (string, bool) {
	// get from cache
	if p.ja3 != nil {
		return p.ja3.GetSNI(), true
	}

	// load from packet
	tcp, ok := p.packet.Layer(layers.LayerTypeTCP).(*layers.TCP)
	if !ok {
		return "", false
	}

	j, err := ja3.ComputeJA3FromSegment(tcp.LayerPayload())
	if err != nil {
		return "", false
	}

	// save to cache
	p.ja3 = j

	return p.ja3.GetSNI(), true
}

func (p *Packet) GetJA3() (string, bool) {
	// get from cache
	if p.ja3 != nil {
		return p.ja3.GetJA3Hash(), true
	}

	// load from packet
	tcp, ok := p.packet.Layer(layers.LayerTypeTCP).(*layers.TCP)
	if !ok {
		return "", false
	}

	j, err := ja3.ComputeJA3FromSegment(tcp.LayerPayload())
	if err != nil {
		return "", false
	}

	// save to cache
	p.ja3 = j

	return p.ja3.GetJA3Hash(), true
}
