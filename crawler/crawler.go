package crawler

import (
	"context"
	"sync"
	"time"

	"github.com/cortze/ragno/db"
	"github.com/cortze/ragno/models"
	"github.com/ethereum/go-ethereum/p2p/enode"
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
	// database
	db *db.PostgresDBService
	// discovery
	peerDisc *peerDisc.PeerDiscovery
	// amount of concurrent dialers
	concurrentDialers int
	// amount of times to retry a connection
	retries int
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
	discv4, err := peerDisc.NewDiscv4(conf.HostPort)
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
		ctx:               ctx,
		doneC:             make(chan struct{}, 1),
		host:              host,
		db:                db,
		concurrentDialers: conf.Dialers,
		retries:           conf.Retries,
		peerDisc:          discvService,
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

	// start workers to connect to peers
	var wgDialers sync.WaitGroup
	logrus.Info("Starting ", c.concurrentDialers, " workers to connect to peers")
	for i := 0; i < c.concurrentDialers; i++ {
		/*
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
		*/
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

func (c *Crawler) Connect(hInfo *models.HostInfo) {
	// try to connect to the peer
	for i := 0; i < c.retries; i++ {
		connAttempt, handsDetails := c.connect(hInfo)
		// track the connection attempt
		c.db.PersistNodeInfo(connAttempt, handsDetails)
		// check if we need to retry
		switch connAttempt.Status {
		case models.FailedConnection:
			// wait for the retry delay
			ticker := time.NewTicker(retryDelay)
			select {
			case <-ticker.C:
				continue
			case <-c.ctx.Done():
				break
			}
		case models.SuccessfulConnection:
			// no need to retry again
			break
		}
	}
}

func (c *Crawler) connect(hInfo *models.HostInfo) (models.ConnectionAttempt, models.NodeInfo) {
	nodeID := enode.PubkeyToIDV4(hInfo.Pubkey)

	connAttempt := models.NewConnectionAttempt(nodeID)
	nInfo, _ := models.NewNodeInfo(nodeID, models.WithHostInfo(*hInfo))

	handshakeDetails, err := c.host.Connect(hInfo)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"node-id": nodeID.String(),
			"error":   err.Error(),
		}).Trace("failed connection")
		connAttempt.Status = models.FailedConnection
	} else {
		logrus.WithFields(logrus.Fields{
			"node-id":      nodeID.String(),
			"client":       handshakeDetails.ClientName,
			"capabilities": handshakeDetails.Capabilities,
		}).Trace("successfull connection")
		connAttempt.Status = models.SuccessfulConnection
		models.WithHandShakeDetails(handshakeDetails)
	}
	connAttempt.Error = err
	connAttempt.Deprecable = false // TODO: upgrade this

	return connAttempt, *nInfo
}
