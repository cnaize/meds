package ip

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gaissmai/bart"

	"github.com/cnaize/meds/lib/util/get"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/types"
)

var _ filter.Filter = (*AbuseIPDB)(nil)

type AbuseIPDB struct {
	*Base

	apiKey     string
	confidence int
}

func NewAbuseIPDB(urls []string, apiKey string, confidence int, logger *logger.Logger, include, exclude *types.IPList) *AbuseIPDB {
	return &AbuseIPDB{
		Base:       NewBase(urls, logger, include, exclude),
		apiKey:     apiKey,
		confidence: confidence,
	}
}

func (f *AbuseIPDB) Name() string {
	return "AbuseIPDB"
}

func (f *AbuseIPDB) Load(ctx context.Context) error {
	defer f.logger.Raw().Info().Str("name", f.Name()).Str("type", string(f.Type())).Msg("Filter loaded")

	return f.Base.Load(ctx)
}

func (f *AbuseIPDB) Update(ctx context.Context) error {
	blocklist := new(bart.Lite)
	for _, u := range f.urls {
		if err := func(u string) error {
			// create params
			params := make(url.Values)
			params.Set("ipVersion", "4")
			params.Set("limit", "9999999")
			params.Set("confidenceMinimum", strconv.Itoa(f.confidence))

			// create request
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s?%s", u, params.Encode()), nil)
			if err != nil {
				return fmt.Errorf("new request: %w", err)
			}

			// add headers
			req.Header.Set("Key", f.apiKey)
			req.Header.Set("Accept", "text/plain")

			// do request
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("do request: %w", err)
			}
			defer resp.Body.Close()

			// check status
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("status code: %d", resp.StatusCode)
			}

			// scan list
			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if len(line) < 1 || strings.HasPrefix(line, "#") {
					continue
				}

				subnet, ok := get.Subnet(line)
				if !ok {
					continue
				}

				blocklist.Insert(subnet)
			}

			return nil
		}(u); err != nil {
			return fmt.Errorf("%s: %w", u, err)
		}
	}

	f.logger.Raw().
		Info().
		Str("name", f.Name()).
		Str("type", string(f.Type())).
		Int("size", blocklist.Size()).
		Msg("Filter updated")
	f.blocklist.Store(blocklist)

	return nil
}
