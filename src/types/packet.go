package types

import (
	"net/netip"

	"github.com/dreadl0ck/ja3"
	"github.com/dreadl0ck/tlsx"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/cnaize/meds/lib/util/get"
)

type tls struct {
	sni string
	ja3 string
}

type Packet struct {
	packet     gopacket.Packet
	asn        ASN
	tls        *tls
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

func (p *Packet) Trusted() bool {
	tcp, ok := p.packet.Layer(layers.LayerTypeTCP).(*layers.TCP)
	if !ok {
		return true
	}

	if len(tcp.Payload) < 1 {
		return false
	}

	return p.tls != nil
}

func (p *Packet) GetProto() (layers.IPProtocol, bool) {
	return get.Proto(p.packet)
}

func (p *Packet) GetSrcIP() (netip.Addr, bool) {
	return get.SrcIP(p.packet)
}

func (p *Packet) GetDstIP() (netip.Addr, bool) {
	return get.DstIP(p.packet)
}

func (p *Packet) GetSrcPort() (uint16, bool) {
	return get.SrcPort(p.packet)
}

func (p *Packet) GetDstPort() (uint16, bool) {
	return get.DstPort(p.packet)
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

// NOTE: pass nil as ASNList to get ASN from cache
func (p *Packet) GetASN(asnlist *ASNList) (ASN, bool) {
	// get from cache
	if asnlist == nil || p.asn.ASN > 0 {
		return p.asn, p.asn.ASN > 0
	}

	srcIP, ok := p.GetSrcIP()
	if !ok {
		return ASN{}, false
	}

	asn, ok := asnlist.Load().Lookup(srcIP)
	if !ok {
		return ASN{}, false
	}

	// save to cache
	p.asn = asn

	return p.asn, true
}

func (p *Packet) GetSNI() (string, bool) {
	if !p.parseTLS() {
		return "", false
	}

	return p.tls.sni, true
}

func (p *Packet) GetJA3() (string, bool) {
	if !p.parseTLS() {
		return "", false
	}

	return p.tls.ja3, true
}

func (p *Packet) parseTLS() bool {
	if p.tls != nil {
		return len(p.tls.sni) > 0 || len(p.tls.ja3) > 0
	}
	p.tls = &tls{}

	tcp, ok := p.packet.Layer(layers.LayerTypeTCP).(*layers.TCP)
	if !ok {
		return false
	}

	var clientHello tlsx.ClientHelloBasic
	if err := clientHello.Unmarshal(tcp.Payload); err != nil {
		return false
	}

	p.tls.sni = clientHello.SNI
	p.tls.ja3 = ja3.DigestHex(&clientHello)

	return true
}
