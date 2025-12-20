package core

import (
	"context"
	"net/netip"

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

	snWhiteList *types.SubnetList
	snBlackList *types.SubnetList
	dmWhiteList *types.DomainList
	dmBlackList *types.DomainList

	filters []filter.Filter
	logger  *logger.Logger
}

func NewWorker(
	subnetWhiteList *types.SubnetList,
	subnetBlackList *types.SubnetList,
	domainWhiteList *types.DomainList,
	domainBlackList *types.DomainList,
	filters []filter.Filter,
	logger *logger.Logger,
) *Worker {
	return &Worker{
		snWhiteList: subnetWhiteList,
		snBlackList: subnetBlackList,
		dmWhiteList: domainWhiteList,
		dmBlackList: domainBlackList,
		filters:     filters,
		logger:      logger,
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
	srcIP, ok := packet.GetSrcIP()
	if !ok {
		w.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
		w.logger.Log(event.NewAccept(zerolog.InfoLevel, "packet skipped", "invalid packet", filter.FilterTypeIP, packet))

		return
	}

	// pass through subnet whitelist
	subnet := netip.PrefixFrom(srcIP, 32)
	if w.snWhiteList.Lookup(subnet) {
		w.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
		w.logger.Log(event.NewAccept(zerolog.InfoLevel, "packet accepted", "whitelisted", filter.FilterTypeIP, packet))

		return
	}

	// pass through subnet blacklist
	if w.snBlackList.Lookup(subnet) {
		w.nfq.SetVerdict(*a.PacketID, nfqueue.NfDrop)
		w.logger.Log(event.NewDrop(zerolog.InfoLevel, "packet dropped", "blacklisted", filter.FilterTypeIP, packet))

		return
	}

	for _, domain := range packet.GetDomains() {
		// pass through domain whitelist
		if w.dmWhiteList.Lookup(domain) {
			w.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
			w.logger.Log(event.NewAccept(zerolog.InfoLevel, "packet accepted", "whitelisted", filter.FilterTypeDomain, packet))

			return
		}

		// pass through domain blacklist
		if w.dmBlackList.Lookup(domain) {
			w.nfq.SetVerdict(*a.PacketID, nfqueue.NfDrop)
			w.logger.Log(event.NewDrop(zerolog.InfoLevel, "packet dropped", "blacklisted", filter.FilterTypeDomain, packet))

			return
		}
	}

	// pass through filters
	for _, filter := range w.filters {
		if !filter.Check(packet) {
			w.nfq.SetVerdict(*a.PacketID, nfqueue.NfDrop)
			w.logger.Log(event.NewDrop(zerolog.InfoLevel, "packet dropped", filter.Name(), filter.Type(), packet))

			return
		}
	}

	// accept by default
	w.nfq.SetVerdict(*a.PacketID, nfqueue.NfAccept)
	w.logger.Log(event.NewAccept(zerolog.DebugLevel, "packet accepted", "default", filter.FilterTypeEmpty, packet))
}
