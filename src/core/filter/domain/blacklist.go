package domain

import (
	"context"
	"slices"

	"github.com/cnaize/meds/src/core/filter"
	"github.com/cnaize/meds/src/core/logger"
	"github.com/cnaize/meds/src/types"
)

var _ filter.Filter = (*BlackList)(nil)

type BlackList struct {
	logger    *logger.Logger
	blacklist *types.DomainList
}

func NewBlackList(logger *logger.Logger, blacklist *types.DomainList) *BlackList {
	return &BlackList{
		logger:    logger,
		blacklist: blacklist,
	}
}

func (f *BlackList) Name() string {
	return filter.FilterNameBlackList
}

func (f *BlackList) Type() filter.FilterType {
	return filter.FilterTypeDomain
}

func (f *BlackList) Load(ctx context.Context) error {
	defer f.logger.Raw().Info().Str("name", f.Name()).Str("type", string(f.Type())).Msg("Filter loaded")

	return nil
}

func (f *BlackList) Check(packet *types.Packet) bool {
	return !slices.ContainsFunc(packet.GetDomains(), f.blacklist.Lookup)
}

func (f *BlackList) Update(ctx context.Context) error {
	return nil
}
