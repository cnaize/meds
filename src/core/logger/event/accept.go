package event

import (
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/rs/zerolog"

	"github.com/cnaize/meds/lib/util/get"
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
	// handle metrics
	defer func() {
		packetsAcceptCounter.WithLabelValues(e.Reason, string(e.Filter)).Inc()
		packetsTotalCounter.Inc()
	}()

	if e.Packet != nil {
		if ip4, ok := e.Packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4); ok {
			var target string
			if e.Filter == filter.FilterTypeDNS {
				target = strings.Join(get.DNSQuestions(e.Packet), ",")
			} else {
				target = ip4.SrcIP.String()
			}

			logger.
				WithLevel(e.Lvl).
				Str("target", target).
				Str("action", string(ActionTypeAccept)).
				Str("reason", e.Reason).
				Str("filter", string(e.Filter)).
				Msg(e.Msg)

			return
		}
	}

	logger.
		WithLevel(e.Lvl).
		Str("target", "empty packet").
		Str("action", string(ActionTypeAccept)).
		Str("reason", e.Reason).
		Str("filter", string(e.Filter)).
		Msg(e.Msg)
}
