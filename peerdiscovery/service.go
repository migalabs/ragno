package peerdiscovery

import (
	"context"
	"strings"
	"sync"

	"github.com/cortze/ragno/db"
	"github.com/cortze/ragno/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Discoverer interface {
	// Run starts the peer discovery process or get the nodes from the file
	Run() (chan *models.ENR, error)
	// Type returns that
	Type() models.DiscoveryType
	// Close the peer discovery
	Close()
}

func StringToDiscoveryType(s string) models.DiscoveryType {
	var discvType models.DiscoveryType = models.UnknownDiscovery
	switch {
	case strings.HasSuffix(s, ".csv"):
		discvType = models.CsvFile
	case s == "discv4":
		discvType = models.Discovery4
	default:
		// do nothing
	}
	return discvType
}

func DiscoveryTypeToString(t models.DiscoveryType) string {
	var discvType = "Unknown"
	switch t {
	case models.CsvFile:
		discvType = "csv-file"
	case models.Discovery4:
		discvType = "discv4"
	default:
		// do nothing
	}
	return discvType
}

// Main peer discovery service that identifies new peers in the network
type PeerDiscovery struct {
	ctx   context.Context
	discv Discoverer
	db    *db.PostgresDBService
	doneC chan struct{}
	wg    sync.WaitGroup
}

func NewPeerDiscovery(ctx context.Context, discv Discoverer, database *db.PostgresDBService) (*PeerDiscovery, error) {
	service := &PeerDiscovery{
		ctx:   ctx,
		discv: discv,
		db:    database,
		doneC: make(chan struct{}),
		wg:    sync.WaitGroup{},
	}
	return service, nil
}

func (d *PeerDiscovery) Run() error {
	newENRs, err := d.discv.Run()
	if err != nil {
		return errors.Wrap(err, "unable to init "+DiscoveryTypeToString(d.discv.Type()))
	}
	d.wg.Add(1)
	go d.discoverPeers(newENRs)

	return nil
}

func (d *PeerDiscovery) discoverPeers(newENRc chan *models.ENR) {
	defer d.wg.Done()
	log := logrus.WithFields(logrus.Fields{
		"discovery-service": DiscoveryTypeToString(d.discv.Type()),
	})
	log.Info("starting peer discovery")

	for {
		select {
		case enr := <-newENRc:
			log.WithField("node-id", enr.ID.String()).Trace("new ENR")
			d.db.PersistENR(enr)

		case <-d.doneC:
			log.Info("shutdown has been triggered")
			return

		case <-d.ctx.Done():
			log.Info("Sudden shutdown detected!")
			return
		}
	}
}

func (d *PeerDiscovery) Close() {
	d.discv.Close()
	d.doneC <- struct{}{}
	d.wg.Wait()
	close(d.doneC)
}
