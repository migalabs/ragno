package models

import (
	"crypto/ecdsa"
	"time"

	"github.com/ethereum/go-ethereum/cmd/devp2p/tooling/ethtest"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/pkg/errors"
)

type NodeInfoOption func(*NodeInfo) error

type ConnectionStatus int8

const (
	NotConnected ConnectionStatus = iota
	FailedConnection
	SuccessfulConnection
)

type NodeInfo struct {
	Timestamp time.Time
	ID        enode.ID
	HostInfo
	HandshakeDetails
	// Metadata Exchange
}

func NewNodeInfo(id enode.ID, opts ...NodeInfoOption) (*NodeInfo, error) {
	nInfo := &NodeInfo{
		Timestamp:        time.Now(),
		ID:               id,
		HostInfo:         HostInfo{},
		HandshakeDetails: HandshakeDetails{},
	}
	for _, opt := range opts {
		err := opt(nInfo)
		if err != nil {
			return nil, err
		}
	}
	return nInfo, nil
}

// WithHostInfo adds the give host info to the NodeInfo struct
func WithHostInfo(hInfo HostInfo) NodeInfoOption {
	return func(n *NodeInfo) error {
		n.HostInfo = hInfo
		return nil
	}
}

// WithHandshakeDetails adds the give handshake info to the NodeInfo struct
func WithHandShakeDetails(d HandshakeDetails) NodeInfoOption {
	return func(n *NodeInfo) error {
		n.HandshakeDetails = d
		return nil
	}
}

func (n *NodeInfo) UpdateTimestamp() {
	n.Timestamp = time.Now()
}

// TODO
// - NodeInfoToCSVrow
// - ReadNodeInfoFromSQLquery

// required info to connect the remote node
type HostInfo struct {
	Pubkey *ecdsa.PublicKey
	IP     string
	TCP    int
}

// Handshake details
type HandshakeDetails struct {
	ClientName   string
	SoftwareInfo uint64
	Capabilities []p2p.Cap
	Error        error
}

func NodeDetailsFromDevp2pHandshake(hdsk ethtest.HandshakeDetails) HandshakeDetails {
	return HandshakeDetails{
		ClientName:   hdsk.ClientName,
		SoftwareInfo: hdsk.SoftwareInfo,
		Capabilities: hdsk.Capabilities,
		Error:        errors.New(string(hdsk.Error)),
	}
}

type ConnectionAttempt struct {
	Timestamp  time.Time
	ID         enode.ID
	Status     ConnectionStatus
	Error      error
	Deprecable bool
}

func NewConnectionAttempt(id enode.ID) ConnectionAttempt {
	return ConnectionAttempt{
		ID:        id,
		Timestamp: time.Now(),
	}
}
