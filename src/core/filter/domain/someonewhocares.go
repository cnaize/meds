package domain

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/armon/go-radix"

	"github.com/cnaize/meds/lib/util/get"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
)

var _ filter.Filter = (*SomeoneWhoCares)(nil)

type SomeoneWhoCares struct {
	*Base
}

func NewSomeoneWhoCares(urls []string, logger *logger.Logger) *SomeoneWhoCares {
	return &SomeoneWhoCares{
		Base: NewBase(urls, logger),
	}
}

func (f *SomeoneWhoCares) Name() string {
	return "SomeoneWhoCares"
}

func (f *SomeoneWhoCares) Load(ctx context.Context) error {
	defer f.logger.Raw().Info().Str("name", f.Name()).Str("type", string(f.Type())).Msg("Filter loaded")

	return f.Base.Load(ctx)
}

func (f *SomeoneWhoCares) Update(ctx context.Context) error {
	blacklist := radix.New()
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
			if len(line) < 1 || strings.HasPrefix(line, "#") {
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

			blacklist.Insert(get.ReversedDomain(domain), struct{}{})
		}
	}

	f.logger.Raw().
		Info().
		Str("name", f.Name()).
		Str("type", string(f.Type())).
		Int("size", blacklist.Len()).
		Msg("Filter updated")
	f.blacklist.Store(blacklist)

	return nil
}
