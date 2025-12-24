package domain

import (
	"context"
	"slices"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/types"
)

var _ filter.Filter = (*WhiteList)(nil)

type WhiteList struct {
	logger    *logger.Logger
	whitelist *types.DomainList
}

func NewWhiteList(logger *logger.Logger, whitelist *types.DomainList) *WhiteList {
	return &WhiteList{
		logger:    logger,
		whitelist: whitelist,
	}
}

func (f *WhiteList) Name() string {
	return filter.FilterNameWhiteList
}

func (f *WhiteList) Type() filter.FilterType {
	return filter.FilterTypeDomain
}

func (f *WhiteList) Load(ctx context.Context) error {
	defer f.logger.Raw().Info().Str("name", f.Name()).Str("type", string(f.Type())).Msg("Filter loaded")

	return nil
}

func (f *WhiteList) Check(packet *types.Packet) bool {
	return slices.ContainsFunc(packet.GetDomains(), f.whitelist.Lookup)
}

func (f *WhiteList) Update(ctx context.Context) error {
	return nil
}
