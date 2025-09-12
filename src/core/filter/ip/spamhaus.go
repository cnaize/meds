package ip

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gaissmai/bart"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
)

var _ filter.Filter = (*Spamhaus)(nil)

type Spamhaus struct {
	*Base
}

func NewSpamhaus(urls []string, logger *logger.Logger) *Spamhaus {
	return &Spamhaus{
		Base: NewBase(urls, logger),
	}
}

func (f *Spamhaus) Name() string {
	return "Spamhaus"
}

func (f *Spamhaus) Update(ctx context.Context) error {
	blackList := new(bart.Lite)
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
			if line == "" || strings.HasPrefix(line, ";") {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) < 1 {
				continue
			}

			prefix, ok := ParsePrefix(fields[0])
			if !ok {
				continue
			}

			blackList.Insert(prefix)
		}
	}

	f.logger.Logger().
		Info().
		Str("name", f.Name()).
		Str("type", string(f.Type())).
		Int("size", blackList.Size()).
		Msg("Filter updated")
	f.blackList.Store(blackList)

	return nil
}
