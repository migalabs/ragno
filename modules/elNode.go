package modules

import (
	"github.com/ethereum/go-ethereum/cmd/devp2p/tooling/ethtest"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"time"
)

type ELNode struct {
	Enode              *enode.Node
	Hinfo              ethtest.HandshakeDetails
	Enr                string
	FirstTimeSeen      time.Time
	LastTimeSeen       time.Time
	FirstTimeConnected time.Time
	LastTimeConnected  time.Time
	LastTimeTried      time.Time
}
