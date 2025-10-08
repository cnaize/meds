package dns

import (
	"context"
	"sync/atomic"

	"github.com/armon/go-radix"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/cnaize/meds/lib/util"
	"github.com/cnaize/meds/lib/util/get"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
)

type Base struct {
	urls      []string
	logger    *logger.Logger
	blacklist atomic.Pointer[radix.Tree]
}

func NewBase(urls []string, logger *logger.Logger) *Base {
	return &Base{
		urls:   urls,
		logger: logger,
	}
}

func (f *Base) Type() filter.FilterType {
	return filter.FilterTypeDNS
}

func (f *Base) Load(ctx context.Context) error {
	f.blacklist.Store(radix.New())

	return nil
}

func (f *Base) Check(packet gopacket.Packet) bool {
	dns, ok := packet.Layer(layers.LayerTypeDNS).(*layers.DNS)
	if !ok {
		return true
	}

	list := f.blacklist.Load()
	// check questions
	for _, question := range dns.Questions {
		if len(question.Name) < 1 {
			continue
		}

		domain := get.ReversedDomain(util.BytesToString(question.Name))
		if _, _, found := list.LongestPrefix(domain); found {
			return false
		}
	}

	// check answers
	for _, answer := range dns.Answers {
		if len(answer.CNAME) < 1 {
			continue
		}

		domain := get.ReversedDomain(util.BytesToString(answer.CNAME))
		if _, _, found := list.LongestPrefix(domain); found {
			return false
		}
	}

	return true
}
