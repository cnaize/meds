package core

import (
	"context"
	"fmt"

	"github.com/appleboy/graceful"
	"github.com/florianl/go-nfqueue"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type Worker struct {
	qnum    uint16
	filters []Filter
	logger  graceful.Logger

	nfq *nfqueue.Nfqueue
}

func NewWorker(qnum uint16, filters []Filter, logger graceful.Logger) *Worker {
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
	w.logger.Infof("packet received")

	// accept empty payload
	if a.Payload == nil {
		w.logger.Infof("empty payload - accept")

		w.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
		return 0
	}

	// WARNING:
	// 1. DON'T MODIFY THE PACKET (NoCopy: true)
	// 2. NOT THREAD SAFE (Lazy: true)
	packet := gopacket.NewPacket(*a.Payload, layers.LayerTypeIPv4, gopacket.DecodeOptions{NoCopy: true, Lazy: true})
	if err := packet.ErrorLayer(); err != nil {
		w.logger.Infof("error occured: %s - accept", err.Error())

		w.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
		return 0
	}

	// pass through filters
	for i, filter := range w.filters {
		if !filter.Check(packet) {
			w.logger.Infof("%d: filter: check failed - block", i)

			w.nfq.SetVerdict(*a.PacketID, nfqueue.NfDrop)
			return 0
		}
	}

	return nfqueue.NfAccept
}

func (w *Worker) errFn(e error) int {
	w.logger.Infof("error received: %s - accept", e.Error())

	return nfqueue.NfAccept
}

func (w *Worker) Close() error {
	if w.nfq != nil {
		return w.nfq.Close()
	}

	return nil
}
