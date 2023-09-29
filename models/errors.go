package models

import (
	"github.com/ethereum/go-ethereum/cmd/devp2p/tooling/ethtest"
	"time"
)

type DialState int8

const (
	// Error States
	ZeroState DialState = iota
	NegativeWithHopeState
	NegativeWithoutHopeState
	PossitiveState

	// Delays based on the the Error State
	ZeroDelay                = 0 * time.Minute
	NegativeWithHopeDelay    = 3 * time.Minute
	NegativeWithoutHopeDalay = 20 * time.Minute
	PossitiveDelay           = 10 * time.Minute
)

func (s DialState) StateToString() (str string) {
	switch s {
	case ZeroState:
		str = "zero"
	case NegativeWithHopeState:
		str = "negative-with-hope"
	case NegativeWithoutHopeState:
		str = "negative-without-hope"
	case PossitiveState:
		str = "possitive"
	default:
		str = "zero"
	}
	return str
}

func (s DialState) DelayFromState() (delay time.Duration) {
	switch s {
	case ZeroState:
		delay = ZeroDelay
	case NegativeWithHopeState:
		delay = NegativeWithHopeDelay
	case NegativeWithoutHopeState:
		delay = NegativeWithoutHopeDalay
	case PossitiveState:
		delay = PossitiveDelay
	default:
		delay = ZeroDelay
	}
	return delay
}

func ParseStateFromError(err error) (state DialState) {
	switch err {
	case ethtest.ErrorNone:
		state = PossitiveState
	case ethtest.ErrorWriteToConnection, ethtest.ErrorEthProtocolNegotiation,
		ethtest.ErrorSnapProtocolNegotiation, ethtest.ErrorBadHandshake:
		state = NegativeWithHopeState
	default:
		state = NegativeWithHopeState
	}
	return state
}
