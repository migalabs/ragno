package modules

import (
	"github.com/ethereum/go-ethereum/cmd/devp2p/tooling/ethtest"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type ELNode struct {
	Enode              *enode.Node
	Hinfo              ethtest.HandshakeDetails
	Enr                string
	FirstTimeSeen      string
	LastTimeSeen       string
	FirstTimeConnected string
	LastTimeConnected  string
	LastTimeTried      string
}
