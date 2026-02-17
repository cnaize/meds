package ip

import (
	"context"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/types"
)

var _ filter.Filter = (*AllowList)(nil)

type AllowList struct {
	logger    *logger.Logger
	allowlist *types.IPList
}

func NewAllowList(logger *logger.Logger, allowlist *types.IPList) *AllowList {
	return &AllowList{
		logger:    logger,
		allowlist: allowlist,
	}
}

func (f *AllowList) Name() string {
	return filter.FilterNameAllowList
}

func (f *AllowList) Type() filter.FilterType {
	return filter.FilterTypeIP
}

func (f *AllowList) Load(ctx context.Context) error {
	defer f.logger.Raw().Info().Str("name", f.Name()).Str("type", string(f.Type())).Msg("Filter loaded")

	return nil
}

func (f *AllowList) Check(packet *types.Packet) bool {
	srcIP, ok := packet.GetSrcIP()
	if !ok {
		return false
	}

	return f.allowlist.Lookup(srcIP)
}

func (f *AllowList) Update(ctx context.Context) error {
	return nil
}
