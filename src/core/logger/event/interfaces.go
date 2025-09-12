package event

import (
	"github.com/rs/zerolog"
)

type Sender interface {
	Send(logger *zerolog.Logger)
}
