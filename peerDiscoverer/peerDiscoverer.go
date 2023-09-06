package peerDiscoverer

import (
	"github.com/cortze/ragno/modules"
)

type PeerDiscoverer interface {
	// Run starts the peer discovery process or get the nodes from the file
	Run(sendingChan chan *modules.ELNode) error
	// sendNodes sends the nodes to the channel
	SendNodes(sendingChan chan *modules.ELNode, node *modules.ELNode)
	// Close the peer discovery
	Close() error
}

type DiscovererType int
