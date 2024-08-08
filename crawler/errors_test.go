package crawler

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestParseConnError(t *testing.T) {
	tests := []struct {
		name     string
		input    error
		expected string
	}{
		{
			name:     "Test None Error",
			input:    errors.New("None"),
			expected: ErrorNone,
		},
		{
			name:     "Test EOF Error",
			input:    errors.New("EOF"),
			expected: ErrorEOF,
		},
		{
			name:     "Test Disconnect Requested Error",
			input:    errors.New("disconnect requested"),
			expected: ErrorDisconnectRequested,
		},
		{
			name:     "Test Decode RLP Disconnect Error",
			input:    errors.New("rlp: expected input list for ethtest.Disconnect"),
			expected: ErrorDecodeRLPdisconnect,
		},
		{
			name:     "Test Useless Peer Error",
			input:    errors.New("useless peer"),
			expected: ErrorUselessPeer,
		},
		{
			name:     "Test Bad Handshake Error",
			input:    errors.New("bad handshake: &ethtest.Error{err:(*errors.errorString)"),
			expected: ErrorBadHandshake,
		},
		{
			name:     "Test Bad Handshake Code 2 Error",
			input:    errors.New("bad status handshake code: 2"),
			expected: ErrorBadHandshake2,
		},
		{
			name:     "Test Timeout Error",
			input:    errors.New("connect: connection timed out"),
			expected: ErrorTimeout,
		},
		{
			name:     "Test Snappy Corrupted Input Error",
			input:    errors.New("snappy: corrupt input"),
			expected: ErrorSnappyCorryptedInput,
		},
		{
			name:     "Test Subprotocol Error",
			input:    errors.New("subprotocol error"),
			expected: ErrorSubprotocol,
		},
		{
			name:     "Test Connection Refused Error",
			input:    errors.New("connect: connection refused"),
			expected: ErrorConnectionRefused,
		},
		{
			name:     "Test Connection Reset Error",
			input:    errors.New("connection reset by peer"),
			expected: ErrorConnectionReset,
		},
		{
			name:     "Test Too Many Peers Error",
			input:    errors.New("too many peers"),
			expected: ErrorTooManyPeers,
		},
		{
			name:     "Test Too Many Peers Error with longer message",
			input:    errors.New("too many peers !?!"),
			expected: ErrorTooManyPeers,
		},
		{
			name:     "Test Unknown Error",
			input:    errors.New("something something..."),
			expected: ErrorUnknown,
		},
	}

	for _, testItem := range tests {
		t.Run(testItem.name, func(t *testing.T) {
			result := ParseConnError(testItem.input)
			require.Equal(t, testItem.expected, result)
		})
	}
}
