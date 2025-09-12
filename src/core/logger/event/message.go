package event

import "github.com/rs/zerolog"

var _ Sender = Message{}

type Message struct {
	Lvl zerolog.Level
	Msg string
}

func NewMessage(lvl zerolog.Level, msg string) Message {
	return Message{
		Lvl: lvl,
		Msg: msg,
	}
}

func (e Message) Send(logger *zerolog.Logger) {
	logger.WithLevel(e.Lvl).Msg(e.Msg)
}
