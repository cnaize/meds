package asn

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/types"
)

var _ filter.Filter = (*Spamhaus)(nil)

type Spamhaus struct {
	*Base
}

func NewSpamhaus(urls []string, logger *logger.Logger, anslist *types.ASNList, include, exclude *types.MapList[uint32]) *Spamhaus {
	return &Spamhaus{
		Base: NewBase(urls, logger, anslist, include, exclude),
	}
}

func (f *Spamhaus) Name() string {
	return "Spamhaus"
}

func (f *Spamhaus) Load(ctx context.Context) error {
	defer f.logger.Raw().Info().Str("name", f.Name()).Str("type", string(f.Type())).Msg("Filter loaded")

	return f.Base.Load(ctx)
}

func (f *Spamhaus) Update(ctx context.Context) error {
	blocklist := make(map[uint32]struct{})
	for _, u := range f.urls {
		if err := func(u string) error {
			// create request
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
			if err != nil {
				return fmt.Errorf("new request: %w", err)
			}

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
				if len(line) < 1 {
					continue
				}

				var entry struct {
					ASN uint32 `json:"asn"`
				}

				if err := json.Unmarshal([]byte(line), &entry); err != nil {
					continue
				}

				blocklist[entry.ASN] = struct{}{}
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
		Int("size", len(blocklist)).
		Msg("Filter updated")
	f.blocklist.Store(&blocklist)

	return nil
}
