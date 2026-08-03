package quarantine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"github.com/maypok86/otter/v2"
	"github.com/nats-io/nats.go"

	"github.com/cnaize/meds/pkg"
	"github.com/cnaize/meds/src/config"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/core/metrics"
	"github.com/cnaize/meds/src/types"
)

var _ filter.Filter = (*QuarantineIP)(nil)

type QuarantineIP struct {
	cfg *config.Config

	nc     *nats.Conn
	logger *logger.Logger

	sub   *nats.Subscription
	cache *otter.Cache[netip.Addr, struct{}]
}

func NewQuarantineIP(cfg *config.Config, nc *nats.Conn, logger *logger.Logger) *QuarantineIP {
	return &QuarantineIP{
		cfg:    cfg,
		nc:     nc,
		logger: logger,
	}
}

func (f *QuarantineIP) Name() string {
	return "Quarantine"
}

func (f *QuarantineIP) Type() filter.FilterType {
	return filter.FilterTypeIP
}

func (f *QuarantineIP) Load(ctx context.Context) error {
	var err error
	f.cache, err = otter.New(
		&otter.Options[netip.Addr, struct{}]{
			MaximumSize:      int(f.cfg.QuarantineIPCacheSize),
			ExpiryCalculator: otter.ExpiryAccessing[netip.Addr, struct{}](f.cfg.QuarantineIPEntityTTL),
			StatsRecorder:    metrics.Get().QuarantineIPCacheStats,
		},
	)
	if err != nil {
		return fmt.Errorf("new cache: %w", err)
	}

	f.sub, err = f.nc.SubscribeSync(pkg.NatsSubjectQuarantineIP + ".*")
	if err != nil {
		return fmt.Errorf("%s: subscribe subject: %w", pkg.NatsSubjectQuarantineIP, err)
	}

	go f.natsHandler(ctx)

	f.logger.Raw().Info().Str("name", f.Name()).Str("type", string(f.Type())).Msg("Filter loaded")

	return nil
}

func (f *QuarantineIP) Check(packet *types.Packet) bool {
	srcIP, ok := packet.GetSrcIP()
	if !ok {
		return true
	}

	if _, found := f.cache.GetIfPresent(srcIP); found {
		return false
	}

	return true
}

func (f *QuarantineIP) Update(ctx context.Context) error {
	return nil
}

func (f *QuarantineIP) natsHandler(ctx context.Context) {
	defer f.sub.Unsubscribe()

	for {
		// read message
		msg, err := f.sub.NextMsgWithContext(ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				f.logger.Raw().Warn().Err(err).Str("name", f.Name()).Str("type", string(f.Type())).Msg("get nats msg failed")
			}
			return
		}

		addr := string(bytes.TrimSpace(msg.Data))

		// parse address
		ip, err := netip.ParseAddr(addr)
		if err != nil {
			f.logger.Raw().Warn().Err(err).Str("name", f.Name()).Str("type", string(f.Type())).Msg("parse nats msg failed")
			continue
		}
		ip = ip.Unmap()

		if !ip.Is4() {
			continue
		}

		// handle subject
		switch msg.Subject {
		case pkg.NatsSubjectQuarantineIPAdd:
			_, set := f.cache.SetIfAbsent(ip, struct{}{})

			// report address
			if f.cfg.FilterAbuseIPDBEnable && f.cfg.FilterAbuseIPDBReportAddr && set {
				go func(ctx context.Context, addr string) {
					ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
					defer cancel()

					if err := f.reportToAbuseIPDB(ctx, addr); err != nil {
						f.logger.Raw().Warn().Err(err).Str("name", f.Name()).Str("type", string(f.Type())).Str("addr", addr).Msg("report to abuseipdb failed")
					}
				}(ctx, addr)
			}
		case pkg.NatsSubjectQuarantineIPDel:
			f.cache.Invalidate(ip)
		}
	}
}

func (f *QuarantineIP) reportToAbuseIPDB(ctx context.Context, addr string) error {
	// create params
	params := make(url.Values)
	params.Set("ip", addr)
	params.Set("categories", "14")
	params.Set("comment", "Malicious activity blocked by Meds firewall")

	// create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.abuseipdb.com/api/v2/report", strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	// add headers
	req.Header.Set("Key", f.cfg.FilterAbuseIPDBApiKey)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// do request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	// check status
	if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusAccepted) {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	return nil
}
