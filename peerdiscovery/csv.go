package peerdiscovery

import (
	csvs "github.com/cortze/ragno/csv"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/cortze/ragno/models"
	"github.com/sirupsen/logrus"
)

type CSV struct {
	csvImporter *csvs.CSV
	enrC        chan *models.ENR
	closeC      chan struct{}
	wg          sync.WaitGroup
}

func NewCSVPeerDiscoverer(file string) (*CSV, error) {
	logrus.Info("Using CSV peer discoverer")

	csvImporter, err := csvs.NewCsvImporter(file)
	if err != nil {
		return nil, err
	}

	disc := &CSV{
		csvImporter: csvImporter,
		enrC:        make(chan *models.ENR),
		closeC:      make(chan struct{}),
	}
	return disc, nil
}

func (c *CSV) Run() (chan *models.ENR, error) {
	// Get all the peers from the csv file
	logrus.Trace("Reading peers from csv file")
	enrSet, err := c.csvImporter.ReadENRset()
	if err != nil {
		return c.enrC, err
	}
	logrus.Debug("Amount of peers read from csv file: ", enrSet.Len())

	closeC := make(chan os.Signal, 1)
	signal.Notify(closeC, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		logrus.Trace("Sending peers to sending channel")
		// send the peers to the sending channel
		enrs := enrSet.GetENRs()
		for _, enr := range enrs {
			select {
			case <-closeC:
				logrus.Info("csvDiscoverer: Shutdown detected")
				return
			default:
			}
			c.notifyNewENR(enr)
		}
		logrus.Trace("csvDiscoverer: Finished sending peers to sending channel")
		return
	}()
	return c.enrC, nil
}

func (c *CSV) notifyNewENR(enr *models.ENR) {
	c.enrC <- enr
}

func (c *CSV) Close() error {
	c.closeC <- struct{}{}
	c.wg.Wait()
	return nil
}

func (c *CSV) Type() models.DiscoveryType {
	return models.CsvFile
}
