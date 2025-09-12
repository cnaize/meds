package event

import (
	"github.com/rs/zerolog"
)

var _ Sender = Error{}

type Error struct {
	Message
	Error error
}

func NewError(lvl zerolog.Level, msg string, err error) Error {
	return Error{
		Message: NewMessage(lvl, msg),
		Error:   err,
	}
}

func (e Error) Send(logger *zerolog.Logger) {
	logger.WithLevel(e.Lvl).Err(e.Error).Msg(e.Msg)
}
