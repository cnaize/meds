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
	w.nfq = nfq
	w.rch = rch

	w.logger.Raw().
		Info().
		Msg("Running worker...")

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
		w.logger.Log(event.NewAccept(zerolog.DebugLevel, "packet skipped", "empty payload", filter.FilterTypeEmpty, nil))

		return
	}

	// accept broken packet
	packet, err := types.NewPacket(*a.Payload)
	if err != nil {
		w.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
		w.logger.Log(event.NewAccept(zerolog.DebugLevel, "packet skipped", "decode failed", filter.FilterTypeEmpty, nil))

		return
	}

	// accept invalid packet
	if _, ok := packet.GetSrcIP(); !ok {
		w.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
		w.logger.Log(event.NewAccept(zerolog.InfoLevel, "packet skipped", "invalid packet", filter.FilterTypeIP, packet))

		return
	}

	// pass through filters
	for _, checker := range w.filters {
		if checker.Check(packet) {
			// accept whitelists
			if checker.Name() == filter.FilterNameWhiteList {
				w.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
				w.logger.Log(event.NewAccept(zerolog.InfoLevel, "packet accepted", checker.Name(), checker.Type(), packet))

				return
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

	// accept by default
	w.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
	w.logger.Log(event.NewAccept(zerolog.DebugLevel, "packet accepted", "default", filter.FilterTypeEmpty, packet))
}
