package event

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/rs/zerolog"

	"github.com/cnaize/meds/src/core/filter"
)

var _ Sender = Packet{}

type Packet struct {
	Message
	Packet     gopacket.Packet
	FilterName string
	FilterType filter.FilterType
}

func NewPacket(lvl zerolog.Level, msg string, packet gopacket.Packet, filterName string, filterType filter.FilterType) Packet {
	return Packet{
		Message:    NewMessage(lvl, msg),
		Packet:     packet,
		FilterName: filterName,
		FilterType: filterType,
	}
}

func (e Packet) Send(logger *zerolog.Logger) {
	ip4, ok := e.Packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	if !ok {
		logger.
			Warn().
			Str("name", e.FilterName).
			Str("type", string(e.FilterType)).
			Msg("empty packet")
		return
	}

	logger.WithLevel(e.Lvl).
		Str("name", e.FilterName).
		Str("type", string(e.FilterType)).
		Str("src_ip", ip4.SrcIP.String()).
		Str("dst_ip", ip4.DstIP.String()).
		Msg(e.Msg)
}
