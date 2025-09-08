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

	nfq, err := nfqueue.Open(&nfqueue.Config{
		NfQueue:      w.qnum,
		MaxQueueLen:  0xFF,
		Copymode:     nfqueue.NfQnlCopyPacket,
		MaxPacketLen: 0xFFFF,
	})
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}

	nfq.RegisterWithErrorFunc(ctx,
		func(a nfqueue.Attribute) int {
			w.logger.Infof("packet received")

			if a.Payload == nil {
				w.logger.Infof("empty payload - accept")

				nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
				return 0
			}

			packet := gopacket.NewPacket(*a.Payload, layers.IPProtocolIPv4, gopacket.NoCopy)
			if err := packet.ErrorLayer(); err != nil {
				w.logger.Infof("error occured: %s - accept", err.Error())

				nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
				return 0
			}

			for _, filter := range w.filters {
				if !filter.Check(packet) {
					w.logger.Infof("check failed - block")

					nfq.SetVerdict(*a.PacketID, nfqueue.NfDrop)
					return 0
				}
			}

			return nfqueue.NfAccept
		},
		func(e error) int {
			w.logger.Infof("error received: %s - accept", e.Error())

			return nfqueue.NfAccept
		},
	)

	w.nfq = nfq

	return nil
}

func (w *Worker) Close() error {
	return w.nfq.Close()
}
