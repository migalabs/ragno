package peerDiscoverer

import (
	"github.com/cortze/ragno/pkg/spec"
)

type Discv4PeerDiscoverer struct {
	sendingChan chan<- *spec.ELNode
	port        int
}

func NewDisv4PeerDiscoverer(conf PeerDiscovererConf) (PeerDiscoverer, error) {
	disc := &Discv4PeerDiscoverer{
		sendingChan: conf.SendingChan,
		port:        conf.Port,
	}
	return disc, nil
}

func (c *Discv4PeerDiscoverer) Run() error {
	return nil
}

func (c *Discv4PeerDiscoverer) sendNodes(node *spec.ELNode) {
	c.sendingChan <- node
}
