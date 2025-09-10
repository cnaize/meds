package ip

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/netip"
	"strings"
	"sync/atomic"

	"github.com/appleboy/graceful"
	"github.com/gaissmai/bart"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/cnaize/meds/src/core/filter"
)

var _ filter.Filter = (*FireHOL)(nil)

type FireHOL struct {
	urls      []string
	logger    graceful.Logger
	blackList atomic.Pointer[bart.Lite]
}

func NewFireHOL(logger graceful.Logger) *FireHOL {
	return &FireHOL{
		urls: []string{
			"https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level1.netset",
			"https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level2.netset",
		},
		logger:    logger,
		blackList: atomic.Pointer[bart.Lite]{},
	}
}

func (f *FireHOL) Load(ctx context.Context) error {
	f.blackList.Store(new(bart.Lite))

	return nil
}

func (f *FireHOL) Check(packet gopacket.Packet) bool {
	ip4, ok := packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
	if !ok {
		f.logger.Errorf("ip: fire hol: invalid packet")
		return true
	}

	list := f.blackList.Load()
	srcIP := netip.AddrFrom4(*(*[4]byte)(ip4.SrcIP.To4()))
	dstIP := netip.AddrFrom4(*(*[4]byte)(ip4.DstIP.To4()))
	if list.Contains(srcIP) || list.Contains(dstIP) {
		f.logger.Infof("ip: fire hol: found: %s -> %s", srcIP.String(), dstIP.String())
		return false
	}

	return true
}

func (f *FireHOL) Update(ctx context.Context) error {
	blackList := new(bart.Lite)
	for _, url := range f.urls {
		// get list
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("%s: get: %w", url, err)
		}
		defer resp.Body.Close()

		// scan list
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			prefix, err := netip.ParsePrefix(line)
			if err != nil {
				ip, err := netip.ParseAddr(line)
				if err != nil {
					continue
				}

				if !ip.Is4() {
					continue
				}

				prefix = netip.PrefixFrom(ip, 32)
			}

			blackList.Insert(prefix)
		}
	}

	f.logger.Infof("Updated: ip filter: FireHOL: size: %d", blackList.Size())
	f.blackList.Store(blackList)

	return nil
}
