package crawler

import (
	"context"
	"sync"
	"time"

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

	// amount of times to retry a connection
	retryAmount int

	// delay between retries (in seconds)
	retryDelay int
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

	// create the peer discoverer
	var discoverer peerDisc.PeerDiscoverer
	if conf.File != "" {
		discoverer, err = peerDisc.NewCSVPeerDiscoverer(conf.File)
	} else {
		discoverer, err = peerDisc.NewDisv4PeerDiscoverer(conf.DiscPort)
	}
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
		retryAmount:       conf.RetryAmount,
		retryDelay:        conf.RetryDelay,
		peerDisc:          discoverer,
	}

	// add all the metrics for each module to the prometheus endp

	return crwl, nil
}

func (c *Crawler) Run() error {

	// channel to receive the peers from the peer discoverer
	connChan := make(chan *modules.ELNode, c.concurrentDialers)

	// start the peer discoverer
	go func() {
		logrus.Info("Starting peer discoverer")
		err := c.peerDisc.Run(connChan)
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
				case peer := <-connChan:
					// try to connect to the peer
					logrus.Trace("Connecting to: ", peer.Enr, " , worker: ", i)
					c.Connect(peer)
					// save the peer
					c.db.PersistNode(*peer)
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

	// try to connect to the peer
	for i := 0; i < c.retryAmount; i++ {
		nodeInfo.Hinfo = c.host.Connect(nodeInfo.Enode)
		if nodeInfo.Hinfo.Error == nil {
			logrus.Trace("Node: ", nodeInfo.Enr, " connected")
			return
		}

		logrus.WithFields(logrus.Fields{
			"retry": i,
			"error": nodeInfo.Hinfo.Error,
		}).Trace("Node: ", nodeInfo.Enr, " failed to connect")

		if i == c.retryAmount-1 || !ShouldRetry(nodeInfo.Hinfo.Error) {
			break
		}
		// wait for the retry delay
		time.Sleep(time.Duration(c.retryDelay) * time.Second)
	}
	logrus.WithFields(logrus.Fields{
		"error": nodeInfo.Hinfo.Error,
	}).Trace("Couldn't connect to node: ", nodeInfo.Enr)
}

func (c *Crawler) Close() {
	// finish discovery

	// stop host

	// stop IP locator

	// stop db

	logrus.Info("Ragno closing routine done! See you!")
}
