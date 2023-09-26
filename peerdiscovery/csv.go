package peerdiscovery

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/cortze/ragno/csv"
	"github.com/cortze/ragno/models"
	"github.com/cortze/ragno/modules"
	"github.com/sirupsen/logrus"
)

type CSV struct {
	csvImporter *csv.CSVImporter
	enrC        chan *models.ENR
}

func NewCSVPeerDiscoverer(file string) (*CSV, error) {
	logrus.Info("Using CSV peer discoverer")

	csvImporter, err := csv.NewCsvImporter(file)
	if err != nil {
		return nil, err
	}

	disc := &CSV{
		csvImporter: csvImporter,
		enrC:        make(chan *models.ENR),
	}
	return disc, nil
}

func (c *CSV) Run(sendingChan chan *modules.ELNode) (chan *models.ENR, error) {
	// Get all the peers from the csv file
	logrus.Trace("Reading peers from csv file")
	peers, err := c.csvImporter.ReadELNodes()
	if err != nil {
		return c.enrC, err
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
			return c.enrC, err
		default:
		}
		c.notifyNewENR(sendingChan, peer)
	}
	logrus.Trace("csvDiscoverer: Finished sending peers to sending channel")

	return c.enrC, nil
}

func (c *CSV) notifyNewENR(sendingChan chan *modules.ELNode, node *modules.ELNode) {
	sendingChan <- node
}

func (c *CSV) Close() error {
	return nil
}
