package crawler

import (
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

func ParseStateFromError(err string) (state DialState) {
	switch err {
	case ErrorNone:
		state = PossitiveState
	case ErrorEOF, ErrorDisconnectRequested, ErrorDecodeRLPdisconnect,
		ErrorBadHandshake, ErrorBadHandshake2, ErrorSnappyCorryptedInput,
		ErrorConnectionReset, ErrorConnectionRefused, ErrorTooManyPeers:
		state = NegativeWithHopeState
	case ErrorTimeout, ErrorUselessPeer:
		state = NegativeWithoutHopeState
	default:
		state = NegativeWithHopeState
	}
	return state
}
