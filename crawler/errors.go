package crawler

import "strings"

type connectErrorType string

const (
	opTimedOut connectErrorType = "operation timed out"

	noMessageReceivedWithin connectErrorType = "no message received within"

	unableToDial connectErrorType = "unable to net.dial node"
	dialFailed   connectErrorType = "dial failed"

	unableToHandshake connectErrorType = "unable to initiate Handshake with node"
	badHandshake      connectErrorType = "bad handshake"
	handshakeFailed   connectErrorType = "handshake failed"

	writeToConnFailed  connectErrorType = "write to connection failed"
	writeToConnFailed2 connectErrorType = "failed to write to connection"
	couldntWriteToConn connectErrorType = "could not write to connection"

	statusExchangeFailed connectErrorType = "status exchange failed"
	badStatusMessage     connectErrorType = "bad status message"
	expectedStatus       connectErrorType = "expected status, got"

	couldntNegoEth  connectErrorType = "could not negotiate eth protocol"
	couldntNegoSnap connectErrorType = "could not negotiate snap protocol"

	wrongHeadBlock       connectErrorType = "wrong head block in status"
	wrongTd              connectErrorType = "wrong TD in status"
	wrongForkId          connectErrorType = "wrong fork ID in status"
	wrongProtocolVersion connectErrorType = "wrong protocol version"

	wrongHeaderAnnounce    connectErrorType = "wrong header in block announcement"
	wrongTDAnnounce        connectErrorType = "wrong TD in announcement"
	wrongBlockHashAnnounce connectErrorType = "wrong block hash in announcement"

	disconnectReceived connectErrorType = "disconnect received"
	expectedDisconnect connectErrorType = "expected disconnect, got"

	protocolVersionNotInConn connectErrorType = "eth protocol version must be set in Conn"

	couldntGetHeaders     connectErrorType = "could not get headers for inbound header request"
	getBlockHeadersFailed connectErrorType = "GetBlockHeader request failed"
	wrongHeader           connectErrorType = "wrong header returned"

	noMessageReceived connectErrorType = "no message received"
	unexpectedMessage connectErrorType = "unexpected message received"

	peeringFailed connectErrorType = "peering failed"

	announceBlockFailed            connectErrorType = "failed to announce block"
	receiveBlockConfirmationFailed connectErrorType = "failed to receive confirmation of block import"

	peeringFailedToStart    connectErrorType = "peering failed"
	createConnectionsFailed connectErrorType = "failed to create connections"

	unexpectedBlockPropagated    connectErrorType = "unexpected: block propagated"
	unexpectedBlockAnnouced      connectErrorType = "unexpected: block announced"
	unexpectedBlockHeadersNumber connectErrorType = "unexpected number of block headers requested"
	unexpectedBlockHeader        connectErrorType = "unexpected block header requested"

	unexpectedNewBlockHashAnnouce connectErrorType = "unexpected new block hash announcement"
	unexpectedBlockHashAnnouce    connectErrorType = "unexpected block hash announcement"

	unexpectedNonEmptyBlock connectErrorType = "unexpected non-empty new block propagated"

	mismatchedHashForNewBlock connectErrorType = "mismatched hash of propagated new block"

	incorrectHeaderReceived connectErrorType = "incorrect header received"

	errorWaitingNodeImportNewBlock connectErrorType = "error waiting for node to import new block"

	unexpectedError connectErrorType = "unexpected error"
	unexpected      connectErrorType = "unexpected:"
)

func (c connectErrorType) String() string {
	return string(c)
}

func (c connectErrorType) Is(err error) bool {
	return strings.Contains(err.Error(), c.String())
}

func ShouldRetry(err error) bool {
	return !(opTimedOut.Is(err) || noMessageReceivedWithin.Is(err))
}
