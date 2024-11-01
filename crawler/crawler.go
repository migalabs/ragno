package crawler

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/cortze/ragno/db"
	peerDisc "github.com/cortze/ragno/peerdiscovery"
	apis "github.com/cortze/ragno/pkg/apis"
	metrics "github.com/cortze/ragno/pkg/metrics"
)

const (
	retryDelay                       = 10 * time.Second
	MetricLoopInterval time.Duration = 15 * time.Second
)

type Crawler struct {
	ctx   context.Context
	doneC chan struct{}
	// host
	host *Host
	// peeringw
	peering *Peering
	// database
	db *db.PostgresDBService
	// IP locator
	IPLocator *apis.IPLocator
	// discovery
	peerDisc *peerDisc.PeerDiscovery
	// metrics
	metrics *metrics.PrometheusMetrics
}

func NewCrawler(ctx context.Context, conf CrawlerRunConf) (*Crawler, error) {
	// create db crawler
	db, err := db.ConnectToDB(ctx, conf.DbEndpoint, conf.Persisters, conf.SnapshotInterval)
	if err != nil {
		logrus.Error("Couldn't init DB")
		return nil, err
	}

	// create a host
	host, err := NewHost(
		ctx,
		conf.HostIP,
		conf.HostPort,
		conf.ConnTimeout,
		// default configuration so far
	)
	if err != nil {
		logrus.Error("failed to create host:")
		return nil, err
	}

	prometheusMetrics := metrics.NewPrometheusMetrics(
		ctx, conf.MetricsIP, conf.MetricsPort, conf.MetricsEndpoint, MetricLoopInterval)

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

	IPLocator := apis.NewIPLocator(ctx, db, conf.IPAPIUrl)

	crwl := &Crawler{
		ctx:      ctx,
		doneC:    make(chan struct{}, 1),
		host:     host,
		peering:  NewPeeringService(ctx, host, db, conf.Dialers, conf.DeprecationTime, IPLocator),
		db:       db,
		peerDisc: discvService,
		metrics:  prometheusMetrics,
		IPLocator: IPLocator,
	}

	crawlerMetricsModule := crwl.GetMetrics()
	prometheusMetrics.AddMetricsModule(crawlerMetricsModule)

	return crwl, nil
}

func (c *Crawler) Run() error {
	// start the peer discoverer
	logrus.Info("Starting peer discoverer")
	err := c.peerDisc.Run()
	if err != nil {
		return errors.Wrap(err, "starting peer-discovery")
	}
	logrus.Info("Starting IP Locator")
	c.IPLocator.Run()
	logrus.Info("Starting metrics")
	c.metrics.Start()
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
	// stop IPLocator
	c.IPLocator.Close()
	// stop metrics
	logrus.Info("crawler: closing metrics")
	c.metrics.Close()
	logrus.Info("Ragno closing routine done! See you!")
}
