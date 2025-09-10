package core

import (
	"context"
	"fmt"

	"github.com/appleboy/graceful"
	"github.com/florianl/go-nfqueue"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/cnaize/meds/src/core/filter"
)

type Worker struct {
	qnum    uint16
	filters []filter.Filter
	logger  graceful.Logger

	nfq *nfqueue.Nfqueue
}

func NewWorker(qnum uint16, filters []filter.Filter, logger graceful.Logger) *Worker {
	return &Worker{
		qnum:    qnum,
		filters: filters,
		logger:  logger,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	w.logger.Infof("Running worker, qnum: %d...", w.qnum)

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
		w.logger.Infof("empty payload -> accept")
		return nfqueue.NfAccept
	}

	// WARNING:
	// 1. DON'T MODIFY THE PACKET (NoCopy: true)
	// 2. NOT THREAD SAFE (Lazy: true)
	packet := gopacket.NewPacket(*a.Payload, layers.LayerTypeIPv4, gopacket.DecodeOptions{NoCopy: true, Lazy: true})
	if err := packet.ErrorLayer(); err != nil {
		w.logger.Infof("decode failed: %s -> accept", err.Error())
		return nfqueue.NfAccept
	}

	// pass through filters
	for _, filter := range w.filters {
		if !filter.Check(packet) {
			return nfqueue.NfDrop
		}
	}

	return nfqueue.NfAccept
}

func (w *Worker) errFn(e error) int {
	w.logger.Infof("received error: %s", e.Error())
	return nfqueue.NfAccept
}

func (w *Worker) Close() error {
	if w.nfq != nil {
		return w.nfq.Close()
	}

	return nil
}
