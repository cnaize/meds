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
			src_ip := ip4.SrcIP.String()
			action := string(ActionTypeAccept)
			reason := e.Reason
			// write message
			logger.
				WithLevel(e.Lvl).
				Str("src_ip", src_ip).
				Str("action", action).
				Str("reason", reason).
				Msg(e.Msg)

			// handle metrics
			packetsAccetCounter.WithLabelValues(src_ip, action, reason)
			packetsTotalCounter.Inc()

			return
		}
	}

	src_ip := "empty packet"
	action := string(ActionTypeAccept)
	reason := e.Reason
	// write message
	logger.
		WithLevel(e.Lvl).
		Str("src_ip", src_ip).
		Str("action", action).
		Str("reason", reason).
		Msg(e.Msg)

	// handle metrics
	packetsAccetCounter.WithLabelValues(src_ip, action, reason)
	packetsTotalCounter.Inc()
}
