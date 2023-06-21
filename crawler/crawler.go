package crawler

import (
	"context"
	"sync"

	"github.com/cortze/ragno/db"
	"github.com/cortze/ragno/modules"
	"github.com/sirupsen/logrus"

	peerDisc "github.com/cortze/ragno/peerDiscoverer"
)

type Crawler struct {
	ctx context.Context

	// host
	host *Host

	// database
	db *db.PostgresDBService

	// discovery
	peerDisc peerDisc.PeerDiscoverer

	// peer connections

	// ip_locator

	// prometheus

	// amount of concurrent dialers
	concurrentDialers int
}

func NewCrawler(ctx context.Context, conf CrawlerRunConf) (*Crawler, error) {
	// create a private key

	// create metrics module

	// create db crawler
	db, err := db.ConnectToDB(ctx, conf.DbEndpoint, conf.ConcurrentSavers)
	if err != nil {
		logrus.Error("Couldn't init DB")
		return nil, err
	}

	// create a host
	host, err := NewHost(
		ctx,
		conf.HostIP,
		conf.HostPort,
		// default configuration so far
	)
	if err != nil {
		logrus.Error("failed to create host:")
		return nil, err
	}

	connChan := make(chan *modules.ELNode, conf.ConcurrentDialers)

	discConf := peerDisc.PeerDiscovererConf{
		SendingChan: connChan,
	}

	// set the file to read the enrs from if provided
	if conf.File != "" {
		discConf.Type = peerDisc.CsvType
		discConf.File = conf.File
	} else {
		discConf.Type = peerDisc.Discv4Type
	}
	discoverer, err := peerDisc.NewPeerDiscoverer(ctx, discConf)
	if err != nil {
		logrus.Error("Couldn't create peer discoverer")
		return nil, err
	}

	// create the discovery modules

	crwl := &Crawler{
		ctx:               ctx,
		host:              host,
		db:                db,
		concurrentDialers: conf.ConcurrentDialers,
		peerDisc:          discoverer,
	}

	// add all the metrics for each module to the prometheus endp

	return crwl, nil
}

func (c *Crawler) Run() error {

	// channel to receive the peers from the peer discoverer
	peerChan := c.peerDisc.Channel()

	// start the peer discoverer
	go func() {
		logrus.Info("Starting peer discoverer")
		err := c.peerDisc.Run()
		if err != nil {
			logrus.Error("Error in peer discoverer: ", err)
		}
	}()

	// start workers to connect to peers
	var wg sync.WaitGroup

	logrus.Info("Starting ", c.concurrentDialers, " workers to connect to peers")
	for i := 0; i < c.concurrentDialers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for {
				select {
				case peer := <-peerChan:
					// try to connect to the peer
					logrus.Trace("Connecting to: ", peer.Enr, " , worker: ", i)
					c.Connect(peer)
					// save the peer
					c.db.Persist(*peer)
				case <-c.ctx.Done():
					return
				}
			}
		}(i)
	}

	// init IP locator

	// init host

	// init discoveries

	return nil
}

func (c *Crawler) Connect(nodeInfo *modules.ELNode) {

	nodeInfo.Hinfo = c.host.Connect(nodeInfo.Enode)
	if nodeInfo.Hinfo.Error != nil {
		logrus.Trace("Node: ", nodeInfo.Enr, ": ", nodeInfo.Hinfo.Error)
	} else {
		logrus.Trace("Node: ", nodeInfo.Enr, " connected")
	}
}

func (c *Crawler) Close() {
	// finish discovery

	// stop host

	// stop IP locator

	// stop db

	logrus.Info("Ragno closing routine done! See you!")
}
