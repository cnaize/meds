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
)

var _ filter.Filter = (*Abuse)(nil)

type Abuse struct {
	*Base
}

func NewAbuse(urls []string, logger *logger.Logger) *Abuse {
	return &Abuse{
		Base: NewBase(urls, logger),
	}
}

func (f *Abuse) Name() string {
	return "Abuse"
}

func (f *Abuse) Load(ctx context.Context) error {
	defer f.logger.Raw().Info().Str("name", f.Name()).Msg("Filter loaded")

	return f.Base.Load(ctx)
}

func (f *Abuse) Update(ctx context.Context) error {
	blacklist := new(bart.Lite)
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
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			prefix, ok := get.NetPrefix(line)
			if !ok {
				continue
			}

			blacklist.Insert(prefix)
		}
	}

	f.logger.Raw().
		Info().
		Str("name", f.Name()).
		Str("type", string(f.Type())).
		Int("size", blacklist.Size()).
		Msg("Filter updated")
	f.blacklist.Store(blacklist)

	return nil
}
