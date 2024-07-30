package models

import (
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/cmd/devp2p/tooling/ethtest"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/forkid"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type NodeInfoOption func(*NodeInfo) error

type ConnectionStatus int8

const EmptyNetworkId = uint64(0)

func (s ConnectionStatus) String() (str string) {
	switch s {
	case NotAttempted:
		str = "not-attempted"
	case FailedConnection:
		str = "failed-attempt"
	case SuccessfulConnection:
		str = "sucessful-attempt"
	default:
		str = "not-attempted"
	}
	return str
}

const (
	NotAttempted ConnectionStatus = iota
	FailedConnection
	SuccessfulConnection
)

type NodeInfo struct {
	Timestamp time.Time
	ID        enode.ID
	HostInfo
	ethtest.HandshakeDetails
	ChainDetails
}

func NewNodeInfo(id enode.ID, opts ...NodeInfoOption) (*NodeInfo, error) {
	nInfo := &NodeInfo{
		Timestamp:        time.Now(),
		ID:               id,
		HostInfo:         HostInfo{},
		HandshakeDetails: ethtest.HandshakeDetails{},
		ChainDetails:     ChainDetails{},
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
func WithHandShakeDetails(d ethtest.HandshakeDetails) NodeInfoOption {
	return func(n *NodeInfo) error {
		n.HandshakeDetails = d
		return nil
	}
}

// WithHostInfo adds the give host info to the NodeInfo struct
func WithChainDetails(cd ChainDetails) NodeInfoOption {
	return func(n *NodeInfo) error {
		n.ChainDetails = cd
		return nil
	}
}

func (n *NodeInfo) UpdateTimestamp() {
	n.Timestamp = time.Now()
}

// required info to connect the remote node
type HostInfo struct {
	ID     enode.ID
	Pubkey *ecdsa.PublicKey
	IP     string
	TCP    int
}

type ConnectionAttempt struct {
	Timestamp  time.Time
	ID         enode.ID
	Status     ConnectionStatus
	Error      string
	Latency    time.Duration
	Deprecable bool
}

func NewConnectionAttempt(id enode.ID) ConnectionAttempt {
	return ConnectionAttempt{
		ID:        id,
		Timestamp: time.Now(),
	}
}

type ChainDetails struct {
	ForkID          forkid.ID
	ProtocolVersion uint32
	HeadHash        common.Hash
	NetworkID       uint64
	TotalDifficulty *big.Int
}

func (d *ChainDetails) IsEmpty() bool {
	return d.NetworkID == EmptyNetworkId
}
