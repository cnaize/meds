package core

import (
	"context"

	"github.com/florianl/go-nfqueue/v2"
	"github.com/rs/zerolog"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/core/logger/event"
	"github.com/cnaize/meds/src/types"
)

type Worker struct {
	nfq *nfqueue.Nfqueue
	rch <-chan nfqueue.Attribute

	filters []filter.Filter
	logger  *logger.Logger
}

func NewWorker(filters []filter.Filter, logger *logger.Logger) *Worker {
	return &Worker{
		filters: filters,
		logger:  logger,
	}
}

func (w *Worker) Run(ctx context.Context, nfq *nfqueue.Nfqueue, rch <-chan nfqueue.Attribute) error {
	w.logger.Raw().Info().Msg("Running worker...")

	w.nfq = nfq
	w.rch = rch

	for {
		select {
		case a := <-w.rch:
			w.handle(a)
		case <-ctx.Done():
			return nil
		}
	}
}

func (w *Worker) handle(a nfqueue.Attribute) {
	// accept empty payload
	if a.Payload == nil {
		w.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
		w.logger.Log(event.NewAccept(zerolog.DebugLevel, "packet accepted", "empty payload", filter.FilterTypeEmpty, nil))

		return
	}

	// accept unknown packet
	packet := types.NewPacket(*a.Payload)
	if _, ok := packet.GetSrcIP(); !ok {
		w.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
		w.logger.Log(event.NewAccept(zerolog.DebugLevel, "packet accepted", "unknown packet", filter.FilterTypeIP, packet))

		return
	}

	// pass through filters
	for _, checker := range w.filters {
		if checker.Check(packet) {
			// accept whitelists
			if checker.Name() == filter.FilterNameWhiteList {
				break
			}
		} else {
			// otherwise drop
			if checker.Name() != filter.FilterNameWhiteList {
				w.nfq.SetVerdict(*a.PacketID, nfqueue.NfDrop)
				w.logger.Log(event.NewDrop(zerolog.InfoLevel, "packet dropped", checker.Name(), checker.Type(), packet))

				return
			}
		}
	}

	// mark trusted connection
	if packet.IsTrusted() {
		w.nfq.SetVerdictWithOption(*a.PacketID, nfqueue.NfAccept, nfqueue.WithConnMark(newMark(a, ConnMark)))
		w.logger.Log(event.NewTrust(zerolog.InfoLevel, "connection marked", "trusted packet", packet))

		return
	}

	// accept by default
	w.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
	w.logger.Log(event.NewAccept(zerolog.DebugLevel, "packet accepted", "default", filter.FilterTypeEmpty, packet))
}

func newMark(a nfqueue.Attribute, mark uint32) uint32 {
	if a.Mark != nil {
		mark |= *a.Mark
	}

	return mark
}
