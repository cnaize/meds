package domain

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/armon/go-radix"
	"github.com/nats-io/nats.go"

	"github.com/cnaize/meds/lib/util/get"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/types"
)

var _ filter.Filter = (*SomeoneWhoCares)(nil)

type SomeoneWhoCares struct {
	*Base
}

func NewSomeoneWhoCares(urls []string, nc *nats.Conn, logger *logger.Logger, include, exclude *types.DomainList) *SomeoneWhoCares {
	return &SomeoneWhoCares{
		Base: NewBase(urls, nc, logger, include, exclude),
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
	blocklist := radix.New()
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

				blocklist.Insert(get.ReversedDomain(domain), struct{}{})
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
		Int("size", blocklist.Len()).
		Msg("Filter updated")
	f.blocklist.Store(blocklist)

	return nil
}
