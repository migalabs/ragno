package crawler

import (
	"github.com/ethereum/go-ethereum/cmd/devp2p/tooling/ethtest"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type NodeToInsert struct {
	Node  *enode.Node
	Info  []string
	Hinfo ethtest.HandshakeDetails
}

func Connect(ctx *cli.Context, node *enode.Node, host *Host, ch chan ethtest.HandshakeDetails) error {

	// no sense for now, but will implement retrying later

	logrus.Info("connecting to: ", node)
	hinfo := host.Connect(node)
	if hinfo.Error != nil {
		logrus.Error(hinfo.Error)
		logrus.Error(`couldn't connect to:`, node.String())
		return hinfo.Error
	}
	logrus.Info("connected to: ", node)
	ch <- hinfo
	return nil
}
