package geo

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gaissmai/bart"

	"github.com/cnaize/meds/lib/util/get"
	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/types"
)

var _ filter.Filter = (*IPLocate)(nil)

type IPLocate struct {
	*Base
}

func NewIPLocate(urls []string, logger *logger.Logger, asnlist *types.ASNList, blacklist *types.CountryList) *IPLocate {
	return &IPLocate{
		Base: NewBase(urls, logger, asnlist, blacklist),
	}
}

func (f *IPLocate) Name() string {
	return "IPLocate"
}

func (f *IPLocate) Load(ctx context.Context) error {
	defer f.logger.Raw().Info().Str("name", f.Name()).Str("type", string(f.Type())).Msg("Filter loaded")

	return f.Base.Load(ctx)
}

func (f *IPLocate) Update(ctx context.Context) error {
	asnlist := new(bart.Table[types.ASN])
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

		// unzip body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read all: %w", err)
		}

		reader := bytes.NewReader(body)
		archive, err := zip.NewReader(reader, reader.Size())
		if err != nil {
			return fmt.Errorf("new zip reader: %w", err)
		}

		for _, file := range archive.File {
			if file.FileInfo().IsDir() {
				continue
			}

			data, err := file.Open()
			if err != nil {
				return fmt.Errorf("%s: open zip file: %w", file.Name, err)
			}

			// scan list
			scanner := bufio.NewScanner(data)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if len(line) < 1 {
					continue
				}

				fields := strings.Split(line, ",")
				if len(fields) < 3 {
					continue
				}

				subnet, ok := get.Subnet(fields[0])
				if !ok {
					continue
				}

				asn, err := strconv.ParseUint(fields[1], 10, 32)
				if err != nil {
					continue
				}

				asnlist.Insert(subnet, types.ASN{
					ASN:     uint32(asn),
					Country: strings.ToLower(fields[2]),
				})
			}
		}
	}

	f.logger.Raw().
		Info().
		Str("name", f.Name()).
		Str("type", string(f.Type())).
		Int("size", asnlist.Size()).
		Msg("Filter updated")
	f.asnlist.Store(asnlist)

	return nil
}
