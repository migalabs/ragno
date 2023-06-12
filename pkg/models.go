package models

import (
	"github.com/ethereum/go-ethereum/cmd/devp2p/tooling/ethtest"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type ELNodeInfo struct {
	Enode         *enode.Node
	Hinfo         ethtest.HandshakeDetails
	Enr           string
	FirstTimeSeen string
	LastTimeSeen  string
}
