package peerDiscoverer

import (
	"github.com/cortze/ragno/modules"
)

type PeerDiscoverer interface {
	// Run starts the peer discovery process or get the nodes from the file
	Run(sendingChan chan *modules.ELNode) error
	// sendNodes sends the nodes to the channel
	sendNodes(sendingChan chan *modules.ELNode, node *modules.ELNode)
}

type DiscovererType int
