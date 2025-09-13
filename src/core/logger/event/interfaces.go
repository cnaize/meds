package event

import (
	"github.com/rs/zerolog"
)

type ActionType string

const (
	ActionTypeAccept ActionType = "accept"
	ActionTypeDrop   ActionType = "drop"
)

type Sender interface {
	Send(logger *zerolog.Logger)
}
