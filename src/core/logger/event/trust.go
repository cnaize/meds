package event

import (
	"github.com/rs/zerolog"

	"github.com/cnaize/meds/src/core/metrics"
	"github.com/cnaize/meds/src/types"
)

var _ Sender = Trust{}

type Trust struct {
	Message

	Reason string
	Packet *types.Packet
}

func NewTrust(lvl zerolog.Level, msg, reason string, packet *types.Packet) Trust {
	return Trust{
		Message: NewMessage(lvl, msg),
		Reason:  reason,
		Packet:  packet,
	}
}

func (e Trust) Send(logger *zerolog.Logger) {
	// handle metrics
	defer func() {
		metrics.Get().TrustConnectionsTotal.WithLabelValues(e.Reason).Inc()
	}()

	if e.Packet != nil {
		var target string
		if srcIP, ok := e.Packet.GetSrcIP(); ok {
			target = srcIP.String()
		}

		logger.
			WithLevel(e.Lvl).
			Str("target", target).
			Str("action", string(ActionTypeTrust)).
			Str("reason", e.Reason).
			Msg(e.Msg)

		return
	}

	logger.
		WithLevel(e.Lvl).
		Str("target", "empty packet").
		Str("action", string(ActionTypeTrust)).
		Str("reason", e.Reason).
		Msg(e.Msg)
}
