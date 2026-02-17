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
			// accept allowlist
			if checker.Name() == filter.FilterNameAllowList {
				mark := addMark(a, ConnMarkTrustList)
				w.nfq.SetVerdictWithOption(
					*a.PacketID,
					nfqueue.NfAccept,
					nfqueue.WithMark(mark),
					nfqueue.WithConnMark(mark),
				)
				w.logger.Log(event.NewAccept(zerolog.InfoLevel, "connection accepted", checker.Name(), checker.Type(), packet))

				return
			}
		} else {
			// drop otherwise
			if checker.Name() != filter.FilterNameAllowList {
				mark := addMark(a, ConnMarkBlockList)
				w.nfq.SetVerdictWithOption(
					*a.PacketID,
					nfqueue.NfRepeat,
					nfqueue.WithMark(mark),
					nfqueue.WithConnMark(mark),
				)
				w.logger.Log(event.NewDrop(zerolog.InfoLevel, "connection dropped", checker.Name(), checker.Type(), packet))

				return
			}
		}
	}

	// accept trusted packet
	if packet.IsTrusted() {
		mark := addMark(a, ConnMarkTrustList)
		w.nfq.SetVerdictWithOption(
			*a.PacketID,
			nfqueue.NfAccept,
			nfqueue.WithMark(mark),
			nfqueue.WithConnMark(mark),
		)
		w.logger.Log(event.NewAccept(zerolog.InfoLevel, "connection trusted", "trusted packet", filter.FilterTypeEmpty, packet))

		return
	}

	// accept by default
	w.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
	w.logger.Log(event.NewAccept(zerolog.DebugLevel, "packet accepted", "default", filter.FilterTypeEmpty, packet))
}

func addMark(a nfqueue.Attribute, mark uint32) uint32 {
	if a.Mark != nil {
		mark |= *a.Mark
	}

	return mark
}
