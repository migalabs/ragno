package crawler

import (
	"context"
	"time"

	"github.com/cortze/ragno/db"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	peerDisc "github.com/cortze/ragno/peerdiscovery"
)

const (
	retryDelay = 10 * time.Second
)

type Crawler struct {
	ctx   context.Context
	doneC chan struct{}
	// host
	host *Host
	// peering
	peering *Peering
	// database
	db *db.PostgresDBService
	// discovery
	peerDisc *peerDisc.PeerDiscovery
}

func NewCrawler(ctx context.Context, conf CrawlerRunConf) (*Crawler, error) {
	// create db crawler
	db, err := db.ConnectToDB(ctx, conf.DbEndpoint, conf.Persisters)
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
	discv4, err := peerDisc.NewDiscv4(ctx, conf.HostPort)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	discvService, err := peerDisc.NewPeerDiscovery(ctx, discv4, db)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	crwl := &Crawler{
		ctx:      ctx,
		doneC:    make(chan struct{}, 1),
		host:     host,
		peering:  NewPeeringService(ctx, host, db, conf.Dialers),
		db:       db,
		peerDisc: discvService,
	}
	return crwl, nil
}

func (c *Crawler) Run() error {
	// start the peer discoverer
	logrus.Info("Starting peer discoverer")
	err := c.peerDisc.Run()
	if err != nil {
		return errors.Wrap(err, "starting peer-discovery")
	}
	// TODO: run metrics
	return c.peering.Run()
}

func (c *Crawler) Close() {
	// finish discovery
	logrus.Info("crawler: closing peer-discovery")
	c.peerDisc.Close()
	// stop peering
	logrus.Info("crawler: closing peering")
	c.peering.Close()
	// close host
	logrus.Info("crawler: closing host")
	c.host.Close()
	// stop db
	logrus.Info("crawler: closing database")
	c.db.Finish()
	logrus.Info("Ragno closing routine done! See you!")
}
