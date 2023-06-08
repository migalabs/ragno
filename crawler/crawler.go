package crawler

import (
	"context"

	"github.com/cortze/ragno/crawler/db"
	"github.com/sirupsen/logrus"
)

type Crawler struct {
	ctx context.Context

	// host
	host *Host

	// database
	db *db.Database

	// discovery

	// peer connections

	// ip_locator

	// prometheus

}

func NewCrawler(ctx context.Context, conf CrawlerRunConf) (*Crawler, error) {
	// create a private key

	// create metrics module

	// create db crawler
	db, err := db.New(ctx, conf.DbEndpoint, 10, 5)
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

	// create the discovery modules

	crwl := &Crawler{
		ctx:  ctx,
		host: host,
		db:   db,
	}

	// add all the metrics for each module to the prometheus endp

	return crwl, nil
}

func (c *Crawler) Run() error {
	// init db

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
