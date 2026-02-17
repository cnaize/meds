package ip

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gaissmai/bart"

	"github.com/cnaize/meds/lib/util/get"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/types"
)

var _ filter.Filter = (*Spamhaus)(nil)

type Spamhaus struct {
	*Base
}

func NewSpamhaus(urls []string, logger *logger.Logger, include, exclude *types.IPList) *Spamhaus {
	return &Spamhaus{
		Base: NewBase(urls, logger, include, exclude),
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
	blocklist := new(bart.Lite)
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
			if len(line) < 1 || strings.HasPrefix(line, ";") {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) < 1 {
				continue
			}

			subnet, ok := get.Subnet(fields[0])
			if !ok {
				continue
			}

			blocklist.Insert(subnet)
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
