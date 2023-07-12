package peerDiscoverer

import (
	"os"
	"os/signal"
	"syscall"

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
	logrus.Debug("Amount of peers read from csv file: ", len(peers))

	closeC := make(chan os.Signal, 1)
	signal.Notify(closeC, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	logrus.Trace("Sending peers to sending channel")
	// send the peers to the sending channel
	for _, peer := range peers {
		select {
		case <-closeC:
			logrus.Info("csvDiscoverer: Shutdown detected")
			return nil
		default:
		}
		c.sendNodes(sendingChan, peer)
	}
	logrus.Trace("csvDiscoverer: Finished sending peers to sending channel")

	return nil
}

func (c *CsvPeerDiscoverer) sendNodes(sendingChan chan *modules.ELNode, node *modules.ELNode) {
	sendingChan <- node
}
