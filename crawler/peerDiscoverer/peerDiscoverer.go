package peerDiscoverer

import (
	"context"

	"github.com/cortze/ragno/pkg/modules"
	"github.com/sirupsen/logrus"
)

type PeerDiscoverer interface {
	// Run starts the peer discovery process or get the nodes from the file
	Run() error
	// sendNodes sends the nodes to the channel
	sendNodes(node *modules.ELNode)
}

type PeerDiscovererConf struct {
	Type        DiscovererType
	SendingChan chan<- *modules.ELNode
	File        string
	Port        int
}

func NewPeerDiscoverer(ctx context.Context, conf PeerDiscovererConf) (PeerDiscoverer, error) {
	switch conf.Type {
	case CsvType:
		logrus.Info("Using CSV peer discoverer")
		return NewCSVPeerDiscoverer(conf)
	case Discv4Type:
		logrus.Info("Using Discv4 peer discoverer")
		return NewDisv4PeerDiscoverer(conf)
	default:
		logrus.Error("Unknown peer discoverer type")
		return nil, nil
	}
}

type DiscovererType int

const (
	Discv4Type DiscovererType = iota
	CsvType
)
