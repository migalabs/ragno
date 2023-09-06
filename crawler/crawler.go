package crawler

import (
	"context"
	"sync"
	"time"

	"github.com/cortze/ragno/db"
	"github.com/cortze/ragno/modules"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	peerDisc "github.com/cortze/ragno/peerDiscoverer"
)

type Crawler struct {
	ctx   context.Context
	doneC chan struct{}
	// host
	host *Host
	// database
	db *db.PostgresDBService
	// discovery
	peerDisc peerDisc.PeerDiscoverer
	// amount of concurrent dialers
	concurrentDialers int
	// amount of times to retry a connection
	retryAmount int
	// delay between retries (in seconds)
	retryDelay int
}

func NewCrawler(ctx context.Context, conf CrawlerRunConf) (*Crawler, error) {
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
		discoverer, err = peerDisc.NewDiscv4(conf.DiscPort)
	}
	if err != nil {
		logrus.Error("Couldn't create peer discoverer")
		return nil, err
	}

	crwl := &Crawler{
		ctx:               ctx,
		doneC:             make(chan struct{}, 1),
		host:              host,
		db:                db,
		concurrentDialers: conf.ConcurrentDialers,
		retryAmount:       conf.RetryAmount,
		retryDelay:        conf.RetryDelay,
		peerDisc:          discoverer,
	}
	return crwl, nil
}

func (c *Crawler) Run() error {
	// channel to receive the peers from the peer discoverer
	connChan := make(chan *modules.ELNode, c.concurrentDialers)

	// start the peer discoverer
	logrus.Info("Starting peer discoverer")
	err := c.peerDisc.Run(connChan)
	if err != nil {
		return errors.Wrap(err, "starting peer-discovery")
	}

	// start workers to connect to peers
	var wgDialers sync.WaitGroup
	logrus.Info("Starting ", c.concurrentDialers, " workers to connect to peers")
	for i := 0; i < c.concurrentDialers; i++ {
		wgDialers.Add(1)
		go func(i int) {
			defer wgDialers.Done()
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

				case <-c.doneC:
					logrus.Info("shutdown detected at worker, closing it")
					return
				}
			}
		}(i)
	}
	logrus.Info("Waiting for dialers to finish")
	wgDialers.Wait()
	close(c.doneC)
	return nil
}

func (c *Crawler) Close() {
	// finish discovery
	c.peerDisc.Close()
	// stop workers
	for i := 0; i < c.concurrentDialers; i++ {
		c.doneC <- struct{}{}
	}
	// close host
	c.host.Close()
	// stop db
	c.db.Finish()

	logrus.Info("Ragno closing routine done! See you!")
}

func (c *Crawler) Connect(nodeInfo *modules.ELNode) {
	// try to connect to the peer
	for i := 0; i < c.retryAmount; i++ {
		nodeInfo.LastTimeTried = time.Now()
		nodeInfo.Hinfo = c.host.Connect(nodeInfo.Enode)
		if nodeInfo.Hinfo.Error == nil {
			logrus.Trace("Node: ", nodeInfo.Enr, " connected")
			nodeInfo.Hinfo.Error = errors.New("None")
			c.db.PersistNode(*nodeInfo)
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
