package crawler

import (
	"context"
	"sync"

	"github.com/cortze/ragno/crawler/db"
	"github.com/cortze/ragno/pkg/spec"
	"github.com/sirupsen/logrus"
)

type Crawler struct {
	ctx context.Context

	// host
	host *Host

	// database
	db *db.PostgresDBService

	// discovery

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

	// set the file to read the enrs from if provided
	if conf.File != "" {
		ctx = context.WithValue(ctx, "File", conf.File)
	}

	// create the discovery modules

	crwl := &Crawler{
		ctx:               ctx,
		host:              host,
		db:                db,
		concurrentDialers: conf.ConcurrentDialers,
	}

	// add all the metrics for each module to the prometheus endp

	return crwl, nil
}

func (c *Crawler) Run() error {
	// init list of peers to connect to
	peers, err := GetListELNodeInfo(c.ctx)
	if err != nil {
		logrus.Error("Couldn't get list of peers")
		return err
	}

	// channel for the peers to connect to
	connChan := make(chan *spec.ELNode, len(peers))

	// fill the channel with the peers
	go func() {
		logrus.Debug("Start fill connChan")
		for _, peer := range peers {
			connChan <- peer
		}
		logrus.Debug("Finish fill connChan")
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

func (c *Crawler) Close() {
	// finish discovery

	// stop host

	// stop IP locator

	// stop db

	logrus.Info("Ragno closing routine done! See you!")
}
