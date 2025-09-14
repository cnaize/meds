package event

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/rs/zerolog"

	"github.com/cnaize/meds/src/core/filter"
)

var _ Sender = Accept{}

type Accept struct {
	Message

	Reason string
	Filter filter.FilterType
	Packet gopacket.Packet
}

func NewAccept(lvl zerolog.Level, msg, reason string, filter filter.FilterType, packet gopacket.Packet) Accept {
	return Accept{
		Message: NewMessage(lvl, msg),
		Reason:  reason,
		Filter:  filter,
		Packet:  packet,
	}
}

func (e Accept) Send(logger *zerolog.Logger) {
	if e.Packet != nil {
		if ip4, ok := e.Packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4); ok {
			src_ip := ip4.SrcIP.String()
			action := string(ActionTypeAccept)
			reason := e.Reason
			filter := string(e.Filter)
			// write message
			logger.
				WithLevel(e.Lvl).
				Str("src_ip", src_ip).
				Str("action", action).
				Str("reason", reason).
				Str("filter", filter).
				Msg(e.Msg)

			// handle metrics
			packetsAccetCounter.WithLabelValues(src_ip, action, reason, filter).Inc()
			packetsTotalCounter.Inc()

			return
		}
	}

	src_ip := "empty packet"
	action := string(ActionTypeAccept)
	reason := e.Reason
	filter := string(e.Filter)
	// write message
	logger.
		WithLevel(e.Lvl).
		Str("src_ip", src_ip).
		Str("action", action).
		Str("reason", reason).
		Str("filter", filter).
		Msg(e.Msg)

	// handle metrics
	packetsAccetCounter.WithLabelValues(src_ip, action, reason, filter).Inc()
	packetsTotalCounter.Inc()
}
