package crawler

import (
	"context"
	"sync"
	"time"

	"github.com/cortze/ragno/db"
	"github.com/cortze/ragno/models"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/sirupsen/logrus"
)

const (
	DeprecationMargin = 3 * time.Hour
	InitDelay         = 2 * time.Second
)

type Peering struct {
	// control
	ctx             context.Context
	appWG           sync.WaitGroup
	orchersterWG    sync.WaitGroup
	dialersWG       sync.WaitGroup
	dialersDoneC    chan struct{}
	orchersterDoneC chan struct{}
	dialC           chan models.HostInfo
	dialers         int

	// necessary services
	host    *Host
	db      *db.PostgresDBService
	nodeSet *NodeOrderedSet
}

func NewPeeringService(ctx context.Context, h *Host, database *db.PostgresDBService, dialers int) *Peering {
	return &Peering{
		ctx:             ctx,
		dialersDoneC:    make(chan struct{}),
		orchersterDoneC: make(chan struct{}),
		dialC:           make(chan models.HostInfo),
		host:            h,
		db:              database,
		nodeSet:         NewNodeOrderedSet(),
		dialers:         dialers,
	}
}

func (p *Peering) Run() error {
	p.appWG.Add(1)
	logrus.Info("running peering service")
	// run dialers
	logrus.Infof("spawning %d peering dialers", p.dialers)
	for workerID := 0; workerID < p.dialers; workerID++ {
		p.dialersWG.Add(1)
		go p.peeringWorker(workerID)
	}
	// run orchester
	p.orchersterWG.Add(1)
	go p.runOrcherster()

	// wait for the process to finish
	p.orchersterWG.Wait()
	for i := 0; i < p.dialers; i++ {
		p.dialersDoneC <- struct{}{}
	}
	p.appWG.Wait()
	return nil
}

func (p *Peering) Close() {
	// trigger the cascade closure starting by the orchester
	p.orchersterDoneC <- struct{}{}
	p.dialersWG.Wait()
	close(p.dialC)
	close(p.dialersDoneC)
	close(p.orchersterDoneC)
	p.appWG.Done() // notify that the dialler has finished
}

func (p *Peering) runOrcherster() {
	logEntry := logrus.WithField("ocherster", 1)
	logEntry.Info("spawning peering dialer orcherster")
	defer func() {
		logEntry.Info("closing peering dial orcherster")
		p.orchersterWG.Done()
	}()
	dialedCache := make(map[enode.ID]struct{})

	startT := time.NewTicker(InitDelay)
	// update the nodes from the db
	updateNodes := func() {
		newNodeSet, err := p.db.GetNonDeprecatedNodes(p.host.localChainStatus.NetworkID)
		if err != nil {
			logEntry.Panic("unable to update local set of nodes from DB")
		}
		p.nodeSet.UpdateSetFromList(newNodeSet)
	}
	updateNodes()
	for {
		// give prior to shut down notifications
		select {
		case <-p.ctx.Done():
			return
		case <-p.orchersterDoneC:
			return
		default:
			if p.nodeSet.IsThereNext() {
				nextNode := p.nodeSet.NextNode()
				_, ok := dialedCache[nextNode.hostInfo.ID]
				if ok {
					continue
				}
				isTimeToDial := nextNode.nextDialTime.Before(time.Now())
				if isTimeToDial {
					p.dialC <- nextNode.hostInfo
				}
				dialedCache[nextNode.hostInfo.ID] = struct{}{}
			} else {
				// still check the contexts in case we have to interrupt
				select {
				case <-p.ctx.Done():
					return
				case <-p.orchersterDoneC:
					return
				case <-startT.C:
				}
				// update the nodeSet
				updateNodes()
				dialedCache = make(map[enode.ID]struct{})
				startT.Reset(InitDelay)
			}
		}
	}
}

func (p *Peering) peeringWorker(workerID int) {
	logEntry := logrus.WithField("dialer-id", workerID)
	logEntry.Debug("spawning peering dialer")
	defer func() {
		logEntry.Info("closing peering dialer")
		p.dialersWG.Done()
	}()
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-p.dialersDoneC:
			return
		case node := <-p.dialC:
			p.Connect(node)
		}
	}
}

// Connect applies the logic of connecting the remote node and persist the necessary results from the attempt
func (p *Peering) Connect(hInfo models.HostInfo) {
	// try to connect to the peer
	connAttempt, nodeInfo, sameNetwork := p.connect(hInfo)
	// handle the result (check if it's deprecable) and update local perception
	p.nodeSet.UpdateNodeFromConnAttempt(hInfo.ID, &connAttempt, sameNetwork)
	// persist the node with all the necessary info
	p.db.PersistNodeInfo(connAttempt, nodeInfo, sameNetwork)
}

// connect offers the low-level connection with the remote peer
func (p *Peering) connect(hInfo models.HostInfo) (models.ConnectionAttempt, models.NodeInfo, bool) {
	logrus.Debug("new node to dial", hInfo.ID.String())
	nodeID := enode.PubkeyToIDV4(hInfo.Pubkey)
	connAttempt := models.NewConnectionAttempt(nodeID)
	nInfo, _ := models.NewNodeInfo(nodeID, models.WithHostInfo(hInfo))

	handshakeDetails, chainDetails, err := p.host.Connect(&hInfo)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"node-id": nodeID.String(),
			"error":   err.Error(),
		}).Debug("failed connection")
		connAttempt.Error = ParseConnError(err)
		connAttempt.Status = models.FailedConnection
	} else {
		logrus.WithFields(logrus.Fields{
			"node-id":          nodeID.String(),
			"client":           handshakeDetails.ClientName,
			"capabilities":     handshakeDetails.Capabilities,
			"network":          chainDetails.NetworkID,
			"fork-id":          chainDetails.ForkID.Hash,
			"head-hash":        chainDetails.HeadHash.String(),
			"protocol-version": chainDetails.ProtocolVersion,
			"total-diff":       chainDetails.TotalDifficulty,
		}).Info("successfull connection")
		connAttempt.Error = ErrorNone
		connAttempt.Status = models.SuccessfulConnection
		nInfo.HandshakeDetails = handshakeDetails
		nInfo.ChainDetails = chainDetails
	}
	return connAttempt, *nInfo, (chainDetails.NetworkID == p.host.localChainStatus.NetworkID)
}
