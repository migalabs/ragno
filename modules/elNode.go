package modules

import (
	"crypto/ecdsa"

	// "github.com/ethereum/go-ethereum/cmd/devp2p/tooling/ethtest"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
)

type NodeInfo struct {
	PeerId       string
	ClientName   string
	Capabilities []p2p.Cap
	SoftwareInfo uint64
}

type NodeControl struct {
	Attempts           int
	SuccessfulAttempts int
	FirstAttempt       string
	LastAttempt        string
	FirstConnection    string
	LastConnection     string
	FirstSeen          string
	LastSeen           string
	LastError          error
	Deprecated         bool
}

type PeerInfo struct {
	IP     []byte
	TCP    int
	UDP    int
	Seq    uint64
	Pubkey ecdsa.PublicKey
	Record enr.Record
}

type ELNode struct {
	NodeId      enode.ID
	Enr         string
	NodeInfo    NodeInfo
	NodeControl NodeControl
	PeerInfo    PeerInfo
}
