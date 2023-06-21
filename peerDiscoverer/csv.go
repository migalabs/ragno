package peerDiscoverer

import (
	"github.com/cortze/ragno/csv"
	"github.com/cortze/ragno/modules"
	"github.com/sirupsen/logrus"
)

type CsvPeerDiscoverer struct {
	csvImporter *csv.CSVImporter
}

func NewCSVPeerDiscoverer(file string) (PeerDiscoverer, error) {
	logrus.Info("Using CSV peer discoverer")

	csvImporter, err := csv.NewCsvImporter(file)
	if err != nil {
		return nil, err
	}

	disc := &CsvPeerDiscoverer{
		csvImporter: csvImporter,
	}
	return disc, nil
}

func (c *CsvPeerDiscoverer) Run(sendingChan chan *modules.ELNode) error {
	// Get all the peers from the csv file
	logrus.Trace("Reading peers from csv file")
	peers, err := c.csvImporter.ReadELNodes()
	if err != nil {
		return err
	}

	logrus.Trace("Sending peers to sending channel")
	// send the peers to the sending channel
	for _, peer := range peers {
		c.sendNodes(sendingChan, peer)
	}
	logrus.Trace("Finished sending peers to sending channel")

	return nil
}

func (c *CsvPeerDiscoverer) sendNodes(sendingChan chan *modules.ELNode, node *modules.ELNode) {
	sendingChan <- node
}
