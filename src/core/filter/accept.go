package filter

import (
	"context"

	"github.com/google/gopacket"

	"github.com/cnaize/meds/src/core"
)

var _ core.Filter = (*Accept)(nil)

type Accept struct {
}

func NewAccept() *Accept {
	return &Accept{}
}

func (f *Accept) Load(ctx context.Context) error {
	return nil
}

func (f *Accept) Check(packet gopacket.Packet) bool {
	return true
}

func (f *Accept) Update(ctx context.Context) error {
	return nil
}
