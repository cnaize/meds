package event

import (
	"strconv"
	"strings"

	"github.com/rs/zerolog"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/metrics"
	"github.com/cnaize/meds/src/types"
)

var _ Sender = Drop{}

type Drop struct {
	Message

	Reason string
	Filter filter.FilterType
	Packet *types.Packet
}

func NewDrop(lvl zerolog.Level, msg, reason string, filter filter.FilterType, packet *types.Packet) Drop {
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
		var target string
		switch e.Filter {
		case filter.FilterTypeIP, filter.FilterTypeRate:
			if srcIP, ok := e.Packet.GetSrcIP(); ok {
				target = srcIP.String()
			}
		case filter.FilterTypeGeo:
			if asn, ok := e.Packet.GetASN(nil); ok {
				target = asn.Country
			}
		case filter.FilterTypeASN:
			if asn, ok := e.Packet.GetASN(nil); ok {
				target = strconv.FormatUint(uint64(asn.ASN), 10)
			}
		case filter.FilterTypeDomain:
			target = strings.Join(e.Packet.GetDomains(), ",")
		case filter.FilterTypeJA3:
			target, _ = e.Packet.GetJA3()
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

	logger.
		WithLevel(e.Lvl).
		Str("target", "empty packet").
		Str("action", string(ActionTypeDrop)).
		Str("reason", e.Reason).
		Str("filter", string(e.Filter)).
		Msg(e.Msg)
}
