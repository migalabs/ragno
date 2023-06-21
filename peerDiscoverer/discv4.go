package peerDiscoverer

import (
	"github.com/cortze/ragno/modules"
)

type Discv4PeerDiscoverer struct {
	sendingChan chan *modules.ELNode
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

func (c *Discv4PeerDiscoverer) Channel() chan *modules.ELNode {
	return c.sendingChan
}

func (c *Discv4PeerDiscoverer) sendNodes(node *modules.ELNode) {
	c.sendingChan <- node
}
