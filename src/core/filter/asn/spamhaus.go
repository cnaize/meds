package asn

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/cnaize/meds/lib/util"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
)

var _ filter.Filter = (*Spamhaus)(nil)

type Spamhaus struct {
	*Base
}

func NewSpamhaus(urls []string, logger *logger.Logger, ipToASN *IPLocate) *Spamhaus {
	return &Spamhaus{
		Base: NewBase(urls, logger, ipToASN),
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
	blacklist := make(map[uint32]bool)
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

			if err := json.Unmarshal(util.StringToBytes(line), &entry); err != nil {
				continue
			}

			if entry.ASN < 1 {
				continue
			}

			blacklist[entry.ASN] = true
		}
	}

	f.logger.Raw().
		Info().
		Str("name", f.Name()).
		Str("type", string(f.Type())).
		Int("size", len(blacklist)).
		Msg("Filter updated")
	f.blacklist.Store(&blacklist)

	return nil
}
