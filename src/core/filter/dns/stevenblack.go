package dns

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/armon/go-radix"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/cnaize/meds/lib/util"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
)

var _ filter.Filter = (*StevenBlack)(nil)

type StevenBlack struct {
	urls      []string
	logger    *logger.Logger
	blackList atomic.Pointer[radix.Tree]
}

func NewStevenBlack(urls []string, logger *logger.Logger) *StevenBlack {
	return &StevenBlack{
		urls:   urls,
		logger: logger,
	}
}

func (f *StevenBlack) Name() string {
	return "StevenBlack"
}

func (f *StevenBlack) Type() filter.FilterType {
	return filter.FilterTypeDNS
}

func (f *StevenBlack) Load(ctx context.Context) error {
	f.blackList.Store(radix.New())

	return nil
}

func (f *StevenBlack) Check(packet gopacket.Packet) bool {
	dns, ok := packet.Layer(layers.LayerTypeDNS).(*layers.DNS)
	if !ok {
		return true
	}

	list := f.blackList.Load()
	for _, question := range dns.Questions {
		domain := normalizeDomain(util.BytesToString(question.Name))
		if _, _, found := list.LongestPrefix(domain); found {
			return false
		}
	}

	return true
}

func (f *StevenBlack) Update(ctx context.Context) error {
	blackList := radix.New()
	for _, url := range f.urls {
		// create request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("%s: new request: %w", url, err)
		}

		// do request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("%s: do request: %w", url, err)
		}
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) < 1 {
				continue
			}

			var domain string
			if len(fields) < 2 {
				domain = fields[0]
			} else {
				domain = fields[1]
			}

			blackList.Insert(normalizeDomain(domain), struct{}{})
		}
	}

	f.logger.Raw().
		Info().
		Str("name", f.Name()).
		Str("type", string(f.Type())).
		Int("size", blackList.Len()).
		Msg("Filter updated")
	f.blackList.Store(blackList)

	return nil
}

func normalizeDomain(domain string) string {
	parts := strings.Split(strings.ToLower(domain), ".")
	slices.Reverse(parts)
	return strings.Join(parts, ".")
}
