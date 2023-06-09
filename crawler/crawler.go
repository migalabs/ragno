package crawler

import (
	"context"
	// "time"

	"github.com/cortze/ragno/crawler/db"
	// "github.com/ethereum/go-ethereum/cmd/devp2p/tooling/ethtest"
	// "github.com/ethereum/go-ethereum/p2p/enode"
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

	// set the file to read the enrs from if provided
	if conf.File != "" {
		ctx = context.WithValue(ctx, "File", conf.File)
	}

	// set the enr to connect to if provided
	if conf.Enr != "" {
		ctx = context.WithValue(ctx, "Enr", conf.Enr)
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
	// init list of peers to connect to
	peers := make([]*ELNodeInfo, 0)

	// get peer if enr is provided
	if c.ctx.Value("Enr") != nil && c.ctx.Value("Enr") != "" {
		sEnr := c.ctx.Value("Enr").(string)
		rEnr := ParseStringToEnr(sEnr)
		peers = append(peers, &ELNodeInfo{
			Enode: rEnr,
			Enr:   sEnr,
		})
		println("peers from enr: ", len(peers))
	}

	// get peers from csv file if provided
	if c.ctx.Value("File") != nil && c.ctx.Value("File") != "" {
		file := c.ctx.Value("File").(string)
		csvImporter, err := NewCsvImporter(file)
		if err != nil {
			logrus.Error("Couldn't create CSV importer")
			return err
		}

		peers, err = ParseCsvToNodeInfo(*csvImporter)
		if err != nil {
			logrus.Error("Couldn't parse the file to Enr")
			return err
		}
	}

	// error if no peers are provided
	if len(peers) == 0 {
		panic("No peers provided")
	}

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
