package event

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/rs/zerolog"
)

var _ Sender = Accept{}

type Accept struct {
	Message

	Reason string
	Packet gopacket.Packet
}

func NewAccept(lvl zerolog.Level, msg, reason string, packet gopacket.Packet) Accept {
	return Accept{
		Message: NewMessage(lvl, msg),
		Reason:  reason,
		Packet:  packet,
	}
}

func (e Accept) Send(logger *zerolog.Logger) {
	if e.Packet != nil {
		if ip4, ok := e.Packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4); ok {
			logger.
				WithLevel(e.Lvl).
				Str("src_ip", ip4.SrcIP.String()).
				Str("action", string(ActionTypeAccept)).
				Str("reason", e.Reason).
				Msg(e.Msg)
			return
		}
	}

	logger.
		WithLevel(e.Lvl).
		Str("src_ip", "empty packet").
		Str("action", string(ActionTypeAccept)).
		Str("reason", e.Reason).
		Msg(e.Msg)
}
