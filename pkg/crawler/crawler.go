package crawler

import (
	"context"

	"github.com/cortze/ragno/pkg/config"

	log "github.com/sirupsen/logrus"
)


type Crawler struct {
	ctx context.Context

	// database

	// discovery

	// peer connections

	// ip_locator

	// prometheus

}


func New(ctx context.Context, conf config.CrawlerRunConf) (*Crawler, error) {
	// create a private key

	// create metrics module

	// create db crawler

	// create a host

	// create the discovery modules 


	crwl := &Crawler{
		ctx: ctx,
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

	log.Info("Ragno closing routine done! See you!")
}
