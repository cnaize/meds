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

var _ filter.Filter = (*AbuseCH)(nil)

type AbuseCH struct {
	*Base
}

func NewAbuseCH(urls []string, logger *logger.Logger, include, exclude *types.IPList) *AbuseCH {
	return &AbuseCH{
		Base: NewBase(urls, logger, include, exclude),
	}
}

func (f *AbuseCH) Name() string {
	return "AbuseCH"
}

func (f *AbuseCH) Load(ctx context.Context) error {
	defer f.logger.Raw().Info().Str("name", f.Name()).Str("type", string(f.Type())).Msg("Filter loaded")

	return f.Base.Load(ctx)
}

func (f *AbuseCH) Update(ctx context.Context) error {
	blocklist := new(bart.Lite)
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
