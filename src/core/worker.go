package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/florianl/go-nfqueue/v2"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/rs/zerolog"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/core/logger/event"
)

type Worker struct {
	qnum    uint16
	filters []filter.Filter
	logger  *logger.Logger

	nfq *nfqueue.Nfqueue
}

func NewWorker(qnum uint16, filters []filter.Filter, logger *logger.Logger) *Worker {
	return &Worker{
		qnum:    qnum,
		filters: filters,
		logger:  logger,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	w.logger.Logger().
		Info().
		Uint16("qnum", w.qnum).
		Msg("Running worker...")

	// open nfqueue
	nfq, err := nfqueue.Open(&nfqueue.Config{
		NfQueue:      w.qnum,
		MaxQueueLen:  0xFF,
		Copymode:     nfqueue.NfQnlCopyPacket,
		MaxPacketLen: 0xFFFF,
	})
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}

	// register nfqueue handlers
	nfq.RegisterWithErrorFunc(ctx, w.hookFn, w.errFn)

	w.nfq = nfq

	return nil
}

func (w *Worker) hookFn(a nfqueue.Attribute) int {
	// accept empty payload
	if a.Payload == nil {
		w.logger.Log(event.NewError(zerolog.WarnLevel, "packet accepted", errors.New("empty payload")))
		return nfqueue.NfAccept
	}

	// WARNING:
	// 1. DON'T MODIFY THE PACKET (NoCopy: true)
	// 2. NOT THREAD SAFE (Lazy: true)
	packet := gopacket.NewPacket(*a.Payload, layers.LayerTypeIPv4, gopacket.DecodeOptions{NoCopy: true, Lazy: true})
	if err := packet.ErrorLayer(); err != nil {
		w.logger.Log(event.NewError(zerolog.WarnLevel, "packet accepted", errors.New("decode failed")))
		return nfqueue.NfAccept
	}

	// pass through filters
	for _, filter := range w.filters {
		if !filter.Check(packet) {
			w.logger.Log(event.NewPacket(zerolog.InfoLevel, "packet dropped", packet, filter.Name(), filter.Type()))
			return nfqueue.NfDrop
		}
	}

	return nfqueue.NfAccept
}

func (w *Worker) errFn(err error) int {
	w.logger.Log(event.NewError(zerolog.ErrorLevel, "error skipped", err))
	return nfqueue.NfAccept
}

func (w *Worker) Close() error {
	if w.nfq != nil {
		return w.nfq.Close()
	}

	return nil
}
