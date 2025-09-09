package filter

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

	"github.com/cnaize/meds/src/core"
)

var _ core.Filter = (*FireHOL)(nil)

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

	// TODO: FIX ME!
	return f.Update(ctx)
}

func (f *FireHOL) Check(packet gopacket.Packet) bool {
	ip4 := packet.Layer(layers.LayerTypeIPv4)
	if ip4 == nil {
		return true
	}
	ip := ip4.(*layers.IPv4)

	list := f.blackList.Load()
	srcIP := netip.AddrFrom4(*(*[4]byte)(ip.SrcIP.To4()))
	dstIP := netip.AddrFrom4(*(*[4]byte)(ip.DstIP.To4()))
	if list.Contains(srcIP) || list.Contains(dstIP) {
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

			if prefix.Addr().IsPrivate() ||
				prefix.Addr().IsLoopback() ||
				prefix.Addr().IsLinkLocalUnicast() ||
				prefix.Addr().IsMulticast() ||
				prefix.Addr().IsUnspecified() {
				continue
			}

			blackList.Insert(prefix)
		}
	}

	f.logger.Infof("Updated FireHOL: size4: %d, size6: %d", blackList.Size4(), blackList.Size6())
	f.blackList.Store(blackList)

	return nil
}
