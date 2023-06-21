package peerDiscoverer

import (
	"github.com/cortze/ragno/csv"
	"github.com/cortze/ragno/modules"
	"github.com/sirupsen/logrus"
)

type CsvPeerDiscoverer struct {
	sendingChan chan *modules.ELNode
	csvImporter *csv.CSVImporter
}

func NewCSVPeerDiscoverer(conf PeerDiscovererConf) (PeerDiscoverer, error) {
	csvImporter, err := csv.NewCsvImporter(conf.File)
	if err != nil {
		return nil, err
	}

	disc := &CsvPeerDiscoverer{
		sendingChan: conf.SendingChan,
		csvImporter: csvImporter,
	}
	return disc, nil
}

func (c *CsvPeerDiscoverer) Run() error {
	// Get all the peers from the csv file
	logrus.Trace("Reading peers from csv file")
	peers, err := c.csvImporter.ReadELNodes()
	if err != nil {
		return err
	}

	logrus.Trace("Sending peers to sending channel")
	// send the peers to the sending channel
	for _, peer := range peers {
		c.sendNodes(peer)
	}
	logrus.Trace("Finished sending peers to sending channel")

	return nil
}

func (c *CsvPeerDiscoverer) Channel() chan *modules.ELNode {
	return c.sendingChan
}

func (c *CsvPeerDiscoverer) sendNodes(node *modules.ELNode) {
	c.sendingChan <- node
}
