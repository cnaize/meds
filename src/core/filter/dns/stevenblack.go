package dns

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/appleboy/graceful"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/cnaize/meds/src/core/filter"
)

var _ filter.Filter = (*StevenBlack)(nil)

type StevenBlack struct {
	urls      []string
	logger    graceful.Logger
	blackList atomic.Pointer[sync.Map]
}

func NewStevenBlack(urls []string, logger graceful.Logger) *StevenBlack {
	return &StevenBlack{
		urls:   urls,
		logger: logger,
	}
}

func (f *StevenBlack) Load(ctx context.Context) error {
	f.blackList.Store(new(sync.Map))

	return nil
}

func (f *StevenBlack) Check(packet gopacket.Packet) bool {
	dns, ok := packet.Layer(layers.LayerTypeDNS).(*layers.DNS)
	if !ok {
		return true
	}

	list := f.blackList.Load()
	for _, question := range dns.Questions {
		name := strings.ToLower(string(question.Name))
		if _, found := list.Load(name); found {
			f.logger.Infof("dns: steven black: found: %s", name)
			return false
		}
	}

	return true
}

func (f *StevenBlack) Update(ctx context.Context) error {
	var count int
	var blackList sync.Map
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

			blackList.Store(strings.ToLower(domain), struct{}{})
			count++
		}
	}

	f.logger.Infof("Updated: dns filter: StevenBlack: size: %d", count)
	f.blackList.Store(&blackList)

	return nil
}
