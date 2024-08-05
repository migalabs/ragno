package crawler

import (
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	ErrorNone    = "none"
	ErrorUnknown = "unknown"

	ErrorEOF                  = "eof"
	ErrorDisconnectRequested  = "disconnect_requested"
	ErrorDecodeRLPdisconnect  = "rlp_decode_disconnect"
	ErrorUselessPeer          = "useless_peer"
	ErrorBadHandshake         = "bad_handshake"
	ErrorBadHandshake2        = "bad_handshake_code_2"
	ErrorTimeout              = "time_out"
	ErrorSnappyCorryptedInput = "snappy_input_corrupted"
	ErrorSubprotocol          = "subprotocol_error"
	ErrorConnectionRefused    = "connection_refused"
	ErrorConnectionReset      = "connection_reset_by_peer"
	ErrorTooManyPeers         = "too_many_peers"
)

var KnownErrors = map[string]string{
	ErrorNone:                 "None",
	ErrorEOF:                  "EOF",
	ErrorDisconnectRequested:  "disconnect requested",
	ErrorDecodeRLPdisconnect:  "rlp: expected input list for ethtest.Disconnect",
	ErrorUselessPeer:          "useless peer",
	ErrorBadHandshake:         "bad handshake: &ethtest.Error{err:(*errors.errorString)",
	ErrorBadHandshake2:        "bad status handshake code: 2",
	ErrorTimeout:              "connect: connection timed out",
	ErrorSnappyCorryptedInput: "snappy: corrupt input",
	ErrorSubprotocol:          "subprotocol error",
	ErrorConnectionRefused:    "connect: connection refused",
	ErrorConnectionReset:      "connection reset by peer",
	ErrorTooManyPeers:         "too many peers",
}

func ParseConnError(err error) string {
	parsedError := ErrorUnknown
	for cleanErr, containable := range KnownErrors {
		if strings.ContainsAny(err.Error(), containable) {
			parsedError = cleanErr
			break
		}
	}
	if parsedError == ErrorUnknown {
		logrus.Warnf("unrecognized error: %s", err.Error())
	}
	return parsedError
}
