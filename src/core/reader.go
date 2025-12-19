package core

import (
	"context"
	"fmt"

	"github.com/florianl/go-nfqueue/v2"
	"github.com/rs/zerolog"

	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/core/logger/event"
)

type Reader struct {
	qnum uint16
	qlen uint32

	logger *logger.Logger

	nfq *nfqueue.Nfqueue
	wch chan nfqueue.Attribute
}

func NewReader(qnum uint16, qlen uint32, logger *logger.Logger) *Reader {
	return &Reader{
		qnum:   qnum,
		qlen:   qlen,
		logger: logger,
		wch:    make(chan nfqueue.Attribute, qlen),
	}
}

func (r *Reader) Run(ctx context.Context) error {
	r.logger.Raw().
		Info().
		Uint16("qnum", r.qnum).
		Msg("Running reader...")

	// open nfqueue
	nfq, err := nfqueue.Open(&nfqueue.Config{
		NfQueue:      r.qnum,
		MaxQueueLen:  r.qlen,
		Copymode:     nfqueue.NfQnlCopyPacket,
		MaxPacketLen: 0xFFFF,
	})
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}

	// register nfqueue handlers
	nfq.RegisterWithErrorFunc(ctx, r.hookFn, r.errFn)

	r.nfq = nfq

	return nil
}

func (r *Reader) Close() error {
	defer close(r.wch)

	if r.nfq != nil {
		return r.nfq.Close()
	}

	return nil
}

func (r *Reader) hookFn(a nfqueue.Attribute) int {
	select {
	case r.wch <- a:
		// good
	default:
		r.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
		r.logger.Log(event.NewError(zerolog.ErrorLevel, "reader chan is full", nil))
	}

	return 0
}

func (r *Reader) errFn(err error) int {
	r.logger.Log(event.NewError(zerolog.ErrorLevel, "nfqueue error skipped", err))

	return 0
}
