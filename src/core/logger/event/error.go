package event

import (
	"github.com/rs/zerolog"

	"github.com/cnaize/meds/src/core/metrics"
)

var _ Sender = Error{}

type Error struct {
	Message

	Err error
}

func NewError(lvl zerolog.Level, msg string, err error) Error {
	return Error{
		Message: NewMessage(lvl, msg),
		Err:     err,
	}
}

func (e Error) Send(logger *zerolog.Logger) {
	defer metrics.Get().ErrorsTotal.WithLabelValues(e.Msg).Inc()

	logger.
		WithLevel(e.Lvl).
		Err(e.Err).
		Msg(e.Msg)
}
