package types

import (
	"net/netip"

	"github.com/dreadl0ck/ja3"
	"github.com/dreadl0ck/tlsx"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/cnaize/meds/lib/util/get"
)

type packetKey uint

const (
	packetKeyASN packetKey = 1 << iota
	packetKeySNI
	packetKeyJA3
	packetKeyTLS
	packetKeyProto
	packetKeySrcIP
	packetKeyDstIP
	packetKeySrcPort
	packetKeyDstPort
	packetKeyDomains
	packetKeyRevDomains
)

type packetTLS struct {
	sni   string
	ja3   string
	hello tlsx.ClientHelloBasic
}

type Packet struct {
	packet gopacket.Packet

	asn        ASN
	tls        packetTLS
	proto      layers.IPProtocol
	srcIP      netip.Addr
	dstIP      netip.Addr
	srcPort    uint16
	dstPort    uint16
	domains    []string
	revDomains []string

	touchMask packetKey
	storeMask packetKey
}

func NewPacket(payload []byte) *Packet {
	// WARNING:
	// 1. DON'T MODIFY PACKET (NoCopy: true)
	// 2. NOT THREAD SAFE (Lazy: true)
	return &Packet{
		packet: gopacket.NewPacket(payload, layers.LayerTypeIPv4,
			gopacket.DecodeOptions{
				NoCopy: true,
				Lazy:   true,
			},
		),
	}
}

func (p *Packet) IsTrusted() bool {
	if _, ok := p.packet.Layer(layers.LayerTypeTCP).(*layers.TCP); ok {
		return p.isStored(packetKeyTLS)
	}

	if _, ok := p.packet.Layer(layers.LayerTypeDNS).(*layers.DNS); ok {
		return false
	}

	return true
}

func (p *Packet) GetProto() (layers.IPProtocol, bool) {
	// check
	if p.isTouched(packetKeyProto) {
		return p.proto, p.isStored(packetKeyProto)
	}

	// touch
	p.touch(packetKeyProto)
	proto, ok := get.Proto(p.packet)
	if !ok {
		return proto, false
	}

	// store
	p.store(packetKeyProto)
	p.proto = proto

	return p.proto, true
}

func (p *Packet) GetSrcIP() (netip.Addr, bool) {
	// check
	if p.isTouched(packetKeySrcIP) {
		return p.srcIP, p.isStored(packetKeySrcIP)
	}

	// touch
	p.touch(packetKeySrcIP)
	srcIP, ok := get.SrcIP(p.packet)
	if !ok {
		return srcIP, false
	}

	// store
	p.store(packetKeySrcIP)
	p.srcIP = srcIP

	return p.srcIP, true
}

func (p *Packet) GetDstIP() (netip.Addr, bool) {
	// check
	if p.isTouched(packetKeyDstIP) {
		return p.dstIP, p.isStored(packetKeyDstIP)
	}

	// touch
	p.touch(packetKeyDstIP)
	dstIP, ok := get.DstIP(p.packet)
	if !ok {
		return dstIP, false
	}

	// store
	p.store(packetKeyDstIP)
	p.dstIP = dstIP

	return p.dstIP, true
}

func (p *Packet) GetSrcPort() (uint16, bool) {
	// check
	if p.isTouched(packetKeySrcPort) {
		return p.srcPort, p.isStored(packetKeySrcPort)
	}

	// touch
	p.touch(packetKeySrcPort)
	srcPort, ok := get.SrcPort(p.packet)
	if !ok {
		return srcPort, false
	}

	// store
	p.store(packetKeySrcPort)
	p.srcPort = srcPort

	return p.srcPort, true
}

func (p *Packet) GetDstPort() (uint16, bool) {
	// check
	if p.isTouched(packetKeyDstPort) {
		return p.dstPort, p.isStored(packetKeyDstPort)
	}

	// touch
	p.touch(packetKeyDstPort)
	dstPort, ok := get.DstPort(p.packet)
	if !ok {
		return dstPort, false
	}

	// store
	p.store(packetKeyDstPort)
	p.dstPort = dstPort

	return p.dstPort, true
}

func (p *Packet) GetDomains() []string {
	// check
	if p.isTouched(packetKeyDomains) {
		return p.domains
	}

	// touch
	p.touch(packetKeyDomains)
	domains := get.DNSDomains(p.packet)
	if sni, ok := p.GetSNI(); ok && len(sni) > 0 {
		domains = append(domains, sni)
	}

	// store
	p.store(packetKeyDomains)
	p.domains = domains

	return p.domains
}

func (p *Packet) GetReversedDomains() []string {
	// check
	if p.isTouched(packetKeyRevDomains) {
		return p.revDomains
	}

	// touch
	p.touch(packetKeyRevDomains)
	domains := p.GetDomains()
	revDomains := make([]string, len(domains))
	for i, domain := range domains {
		revDomains[i] = get.ReversedDomain(domain)
	}

	// store
	p.store(packetKeyRevDomains)
	p.revDomains = revDomains

	return p.revDomains
}

// NOTE: pass nil as ASNList to get ASN from cache
func (p *Packet) GetASN(asnlist *ASNList) (ASN, bool) {
	// check
	if asnlist == nil || p.isTouched(packetKeyASN) {
		return p.asn, p.isStored(packetKeyASN)
	}

	// touch
	p.touch(packetKeyASN)
	srcIP, ok := p.GetSrcIP()
	if !ok {
		return p.asn, false
	}

	asn, ok := asnlist.Load().Lookup(srcIP)
	if !ok {
		return p.asn, false
	}

	// store
	p.store(packetKeyASN)
	p.asn = asn

	return p.asn, true
}

func (p *Packet) GetSNI() (string, bool) {
	// check
	if p.isTouched(packetKeySNI) {
		return p.tls.sni, p.isStored(packetKeySNI)
	}

	// touch
	p.touch(packetKeySNI)
	hello, ok := p.parseTLS()
	if !ok {
		return p.tls.sni, false
	}

	// store
	p.store(packetKeySNI)
	p.tls.sni = hello.SNI

	return p.tls.sni, true
}

func (p *Packet) GetJA3() (string, bool) {
	// check
	if p.isTouched(packetKeyJA3) {
		return p.tls.ja3, p.isStored(packetKeyJA3)
	}

	// touch
	p.touch(packetKeyJA3)
	hello, ok := p.parseTLS()
	if !ok {
		return p.tls.ja3, false
	}

	// store
	p.store(packetKeyJA3)
	p.tls.ja3 = ja3.DigestHex(&hello)

	return p.tls.ja3, true
}

func (p *Packet) parseTLS() (tlsx.ClientHelloBasic, bool) {
	// check
	if p.isTouched(packetKeyTLS) {
		return p.tls.hello, p.isStored(packetKeyTLS)
	}

	// touch
	p.touch(packetKeyTLS)
	tcp, ok := p.packet.Layer(layers.LayerTypeTCP).(*layers.TCP)
	if !ok {
		return p.tls.hello, false
	}

	var clientHello tlsx.ClientHelloBasic
	if err := clientHello.Unmarshal(tcp.Payload); err != nil {
		return p.tls.hello, false
	}

	// store
	p.store(packetKeyTLS)
	p.tls.hello = clientHello

	return p.tls.hello, true
}

func (p *Packet) touch(key packetKey) {
	p.touchMask |= key
}

func (p *Packet) isTouched(key packetKey) bool {
	return p.touchMask&key != 0
}

func (p *Packet) store(key packetKey) {
	p.storeMask |= key
}

func (p *Packet) isStored(key packetKey) bool {
	return p.storeMask&key != 0
}
