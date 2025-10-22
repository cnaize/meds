package core

import (
	"context"
	"fmt"
	"net/netip"
	"slices"

	"github.com/florianl/go-nfqueue/v2"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/rs/zerolog"

	"github.com/cnaize/meds/lib/util/get"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/core/logger/event"
	"github.com/cnaize/meds/src/types"
)

type Worker struct {
	qnum uint16

	snWhiteList *types.SubnetList
	snBlackList *types.SubnetList
	dmWhiteList *types.DomainList
	dmBlackList *types.DomainList

	filters []filter.Filter
	logger  *logger.Logger

	nfq *nfqueue.Nfqueue
}

func NewWorker(
	qnum uint16,
	subnetWhiteList *types.SubnetList,
	subnetBlackList *types.SubnetList,
	domainWhiteList *types.DomainList,
	domainBlackList *types.DomainList,
	filters []filter.Filter,
	logger *logger.Logger,
) *Worker {
	return &Worker{
		qnum:        qnum,
		snWhiteList: subnetWhiteList,
		snBlackList: subnetBlackList,
		dmWhiteList: domainWhiteList,
		dmBlackList: domainBlackList,
		filters:     filters,
		logger:      logger,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	w.logger.Raw().
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
		w.logger.Log(event.NewAccept(zerolog.DebugLevel, "packet skipped", "empty payload", filter.FilterTypeEmpty, nil))

		w.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
		return 0
	}

	// WARNING:
	// 1. DON'T MODIFY THE PACKET (NoCopy: true)
	// 2. NOT THREAD SAFE (Lazy: true)
	packet := gopacket.NewPacket(*a.Payload, layers.LayerTypeIPv4, gopacket.DecodeOptions{NoCopy: true, Lazy: true})
	if err := packet.ErrorLayer(); err != nil {
		w.logger.Log(event.NewAccept(zerolog.DebugLevel, "packet skipped", "decode failed", filter.FilterTypeEmpty, nil))

		w.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
		return 0
	}

	// accept invalid packet
	srcIP, ok := get.PacketSrcIP(packet)
	if !ok {
		w.logger.Log(event.NewAccept(zerolog.InfoLevel, "packet skipped", "invalid packet", filter.FilterTypeIP, packet))

		w.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
		return 0
	}

	// pass through subnet whitelist
	subnet := netip.PrefixFrom(srcIP, 32)
	if w.snWhiteList.Lookup(subnet) {
		w.logger.Log(event.NewAccept(zerolog.InfoLevel, "packet accepted", "whitelisted", filter.FilterTypeIP, packet))

		w.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
		return 0
	}

	// pass through domain whitelist
	domains := get.Domains(packet)
	if slices.ContainsFunc(domains, w.dmWhiteList.Lookup) {
		w.logger.Log(event.NewAccept(zerolog.InfoLevel, "packet accepted", "whitelisted", filter.FilterTypeDNS, packet))

		w.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
		return 0
	}

	// pass through filters
	for _, filter := range w.filters {
		if !filter.Check(packet) {
			w.logger.Log(event.NewDrop(zerolog.InfoLevel, "packet dropped", filter.Name(), filter.Type(), packet))

			w.nfq.SetVerdict(*a.PacketID, nfqueue.NfDrop)
			return 0
		}
	}

	// pass through subnet blacklist
	if w.snBlackList.Lookup(subnet) {
		w.logger.Log(event.NewDrop(zerolog.InfoLevel, "packet dropped", "blacklisted", filter.FilterTypeIP, packet))

		w.nfq.SetVerdict(*a.PacketID, nfqueue.NfDrop)
		return 0
	}

	// pass through domain blacklist
	if slices.ContainsFunc(domains, w.dmBlackList.Lookup) {
		w.logger.Log(event.NewDrop(zerolog.InfoLevel, "packet dropped", "blacklisted", filter.FilterTypeDNS, packet))

		w.nfq.SetVerdict(*a.PacketID, nfqueue.NfDrop)
		return 0
	}

	// accept by default
	w.logger.Log(event.NewAccept(zerolog.DebugLevel, "packet accepted", "default", filter.FilterTypeEmpty, packet))

	w.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
	return 0
}

func (w *Worker) errFn(err error) int {
	w.logger.Log(event.NewError(zerolog.ErrorLevel, "error skipped", err))
	return 0
}

func (w *Worker) Close() error {
	if w.nfq != nil {
		return w.nfq.Close()
	}

	return nil
}
