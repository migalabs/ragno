package peerDiscoverer

import (
	"github.com/cortze/ragno/modules"
	"github.com/sirupsen/logrus"
)

type Discv4PeerDiscoverer struct {
	port int
}

func NewDisv4PeerDiscoverer(port int) (PeerDiscoverer, error) {
	logrus.Info("Using Discv4 peer discoverer")

	disc := &Discv4PeerDiscoverer{
		port: port,
	}
	return disc, nil
}

func (c *Discv4PeerDiscoverer) Run(sendingChan chan *modules.ELNode) error {
	return nil
}

func (c *Discv4PeerDiscoverer) sendNodes(sendingChan chan *modules.ELNode, node *modules.ELNode) {
	sendingChan <- node
}
