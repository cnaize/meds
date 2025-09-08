package filter

import (
	"context"

	"github.com/google/gopacket"
)

type True struct {
}

func NewTrue() *True {
	return &True{}
}

func (c *True) Check(packet gopacket.Packet) bool {
	return true
}

func (c *True) Update(ctx context.Context) error {
	return nil
}
