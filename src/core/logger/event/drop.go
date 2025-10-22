package event

import (
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/rs/zerolog"

	"github.com/cnaize/meds/lib/util/get"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/metrics"
)

var _ Sender = Drop{}

type Drop struct {
	Message

	Reason string
	Filter filter.FilterType
	Packet gopacket.Packet
}

func NewDrop(lvl zerolog.Level, msg, reason string, filter filter.FilterType, packet gopacket.Packet) Drop {
	return Drop{
		Message: NewMessage(lvl, msg),
		Reason:  reason,
		Filter:  filter,
		Packet:  packet,
	}
}

func (e Drop) Send(logger *zerolog.Logger) {
	// handle metrics
	defer func() {
		metrics.Get().PacketsDroppedTotal.WithLabelValues(e.Reason, string(e.Filter)).Inc()
		metrics.Get().PacketsProcessedTotal.Inc()
	}()

	if e.Packet != nil {
		if ip4, ok := e.Packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4); ok {
			var target string
			switch e.Filter {
			case filter.FilterTypeRate, filter.FilterTypeIP:
				target = ip4.SrcIP.String()
			case filter.FilterTypeDNS:
				target = strings.Join(get.Domains(e.Packet), ",")
			}

			logger.
				WithLevel(e.Lvl).
				Str("target", target).
				Str("action", string(ActionTypeDrop)).
				Str("reason", e.Reason).
				Str("filter", string(e.Filter)).
				Msg(e.Msg)

			return
		}
	}

	logger.
		WithLevel(e.Lvl).
		Str("target", "empty packet").
		Str("action", string(ActionTypeDrop)).
		Str("reason", e.Reason).
		Str("filter", string(e.Filter)).
		Msg(e.Msg)
}
