package crawler

import (
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	ErrorNone    = "none"
	ErrorUnknown = "unknown"

	ErrorEOF                    = "eof"
	ErrorDisconnectRequested    = "disconnect_requested"
	ErrorDecodeRLPdisconnect    = "rlp_decode_disconnect"
	ErrorUselessPeer            = "useless_peer"
	ErrorBadHandshake           = "bad_handshake"
	ErrorBadHandshake2          = "bad_handshake_code_2"
	ErrorTimeout                = "time_out"
	ErrorSnappyCorruptedInput   = "snappy_input_corrupted"
	ErrorSubprotocol            = "subprotocol_error"
	ErrorConnectionRefused      = "connection_refused"
	ErrorConnectionReset        = "connection_reset_by_peer"
	ErrorTooManyPeers           = "too_many_peers"
	ErrorIOTimeout              = "i/o_timeout"
	ErrorNoRouteToHost          = "no_route_to_host"
	ErrorProtocolNegotiation    = "eth_protocols_negotiation"
	ErrorBadHandshakeDisconnect = "bad_handshake_disconnect"
)

var KnownErrors = map[string]string{
	ErrorNone:                   "None",
	ErrorEOF:                    "EOF",
	ErrorDisconnectRequested:    "disconnect requested",
	ErrorDecodeRLPdisconnect:    "rlp: expected input list for ethtest.Disconnect",
	ErrorUselessPeer:            "useless peer",
	ErrorBadHandshake:           "bad handshake: ",
	ErrorBadHandshake2:          "bad status handshake code: 2",
	ErrorTimeout:                "connect: connection timed out",
	ErrorSnappyCorruptedInput:   "snappy: corrupt input",
	ErrorSubprotocol:            "subprotocol error",
	ErrorConnectionRefused:      "connect: connection refused",
	ErrorConnectionReset:        "connection reset by peer",
	ErrorTooManyPeers:           "too many peers",
	ErrorIOTimeout:              "i/o timeout",
	ErrorNoRouteToHost:          "connect: no route to host",
	ErrorProtocolNegotiation:    "eth protocols negotiation",
	ErrorBadHandshakeDisconnect: "bad status handshake disconnect: ",
}

func ParseConnError(err error) string {
	parsedError := ErrorUnknown
	for cleanErr, containable := range KnownErrors {
		if strings.Contains(err.Error(), containable) {
			parsedError = cleanErr
			break
		}
	}
	if parsedError == ErrorUnknown {
		logrus.Warnf("unrecognized error: %s", err.Error())
	}
	return parsedError
}
