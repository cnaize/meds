package logger

import (
	"github.com/rs/zerolog"

	"github.com/cnaize/meds/src/core/logger/event"
)

type Logger struct {
	logger *zerolog.Logger
	events chan event.Sender
}

func NewLogger(logger *zerolog.Logger) *Logger {
	return &Logger{
		logger: logger,
		events: make(chan event.Sender, 256),
	}
}

func (l *Logger) Run(workers uint) {
	for range workers {
		go l.recvLoop()
	}
}

func (l *Logger) Log(e event.Sender) {
	select {
	case l.events <- e:
	default:
		l.logger.Warn().Msgf("event dropped: %T", e)
	}
}

func (l *Logger) Logger() *zerolog.Logger {
	return l.logger
}

func (l *Logger) recvLoop() {
	for e := range l.events {
		e.Send(l.logger)
	}
}
