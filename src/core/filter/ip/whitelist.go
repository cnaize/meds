package ip

import (
	"context"
	"net/netip"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/types"
)

var _ filter.Filter = (*WhiteList)(nil)

type WhiteList struct {
	logger    *logger.Logger
	whitelist *types.SubnetList
}

func NewWhiteList(logger *logger.Logger, whitelist *types.SubnetList) *WhiteList {
	return &WhiteList{
		logger:    logger,
		whitelist: whitelist,
	}
}

func (f *WhiteList) Name() string {
	return filter.FilterNameWhiteList
}

func (f *WhiteList) Type() filter.FilterType {
	return filter.FilterTypeIP
}

func (f *WhiteList) Load(ctx context.Context) error {
	defer f.logger.Raw().Info().Str("name", f.Name()).Str("type", string(f.Type())).Msg("Filter loaded")

	return nil
}

func (f *WhiteList) Check(packet *types.Packet) bool {
	srcIP, ok := packet.GetSrcIP()
	if !ok {
		return false
	}

	return f.whitelist.Lookup(netip.PrefixFrom(srcIP, 32))
}

func (f *WhiteList) Update(ctx context.Context) error {
	return nil
}
