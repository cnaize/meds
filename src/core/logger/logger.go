package logger

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/cnaize/meds/src/core/logger/event"
)

type Logger struct {
	logger *zerolog.Logger
	events chan event.Sender
}

func NewLogger(logger *zerolog.Logger, qlen uint) *Logger {
	return &Logger{
		logger: logger,
		events: make(chan event.Sender, qlen),
	}
}

func (l *Logger) Raw() *zerolog.Logger {
	return l.logger
}

func (l *Logger) Run(ctx context.Context, workers uint) {
	for range workers {
		go l.sendLoop(ctx)
	}
}

func (l *Logger) Log(e event.Sender) {
	select {
	case l.events <- e:
	default:
		l.logger.Warn().Msgf("event dropped: %T", e)
	}
}

func (l *Logger) sendLoop(ctx context.Context) {
	for {
		select {
		case e := <-l.events:
			e.Send(l.logger)
		case <-ctx.Done():
			return
		}
	}
}
