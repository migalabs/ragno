package crawler

import (
	"context"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/sirupsen/logrus"

	"github.com/cortze/ragno/db"
	"github.com/cortze/ragno/models"
	"github.com/cortze/ragno/pkg/apis"
)

const (
	DeprecationMargin = 48 * time.Hour
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
	host      *Host
	db        *db.PostgresDBService
	nodeSet   *NodeOrderedSet
	IPLocator *apis.IPLocator
}

func NewPeeringService(
	ctx context.Context, h *Host, database *db.PostgresDBService, dialers int,
	IPLocator *apis.IPLocator,
) *Peering {
	return &Peering{
		ctx:             ctx,
		dialersDoneC:    make(chan struct{}),
		orchersterDoneC: make(chan struct{}),
		dialC:           make(chan models.HostInfo),
		host:            h,
		db:              database,
		nodeSet:         NewNodeOrderedSet(),
		dialers:         dialers,
		IPLocator:       IPLocator,
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
			logEntry.Panic(err.Error())
			logEntry.Panic("unable to update local set of nodes from DB")
		}
		p.nodeSet.UpdateListFromSet(newNodeSet)
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
	// If the current node's IP is public, locate IP info and persist if valid
	p.requestIPInfo(hInfo.ID.String(), hInfo.IP)
}

// connect offers the low-level connection with the remote peer
func (p *Peering) connect(hInfo models.HostInfo) (models.ConnectionAttempt, models.NodeInfo, bool) {
	logrus.Debug("new node to dial", hInfo.ID.String())
	nodeID := enode.PubkeyToIDV4(hInfo.Pubkey)
	connAttempt := models.NewConnectionAttempt(nodeID)
	nInfo, _ := models.NewNodeInfo(nodeID, models.WithHostInfo(hInfo))

	t := time.Now()
	handshakeDetails, chainDetails, err := p.host.Connect(&hInfo)
	RTT := time.Since(t)

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
		connAttempt.Latency = RTT
		nInfo.HandshakeDetails = handshakeDetails
		nInfo.ChainDetails = chainDetails
	}
	return connAttempt, *nInfo, (chainDetails.NetworkID == p.host.localChainStatus.NetworkID)
}

func (p *Peering) requestIPInfo(nodeID string, IP string) {
	if models.IsIPPublic(net.ParseIP(IP)) {
		// get location from the received peer
		p.IPLocator.LocateIP(IP)
	} else {
		logrus.Warnf("new peer %s had a non-public IP %s", nodeID, IP)
	}
}

// --- Ordered Set ---

// List of Nodes ordered by time for next connection
type NodeOrderedSet struct {
	m        sync.RWMutex
	nodePtr  int
	nodeList []*QueuedNode
	nodeMap  map[string]*QueuedNode
}

func NewNodeOrderedSet() *NodeOrderedSet {
	return &NodeOrderedSet{
		nodePtr:  0,
		nodeList: make([]*QueuedNode, 0),
		nodeMap:  make(map[string]*QueuedNode),
	}
}

func (s *NodeOrderedSet) UpdateListFromSet(nSet []models.HostInfo) {
	// try to add the missing nodes from the DB into the NodeSet
	newNodes := 0
	for _, newNode := range nSet {
		exists := s.IsPeerAlready(newNode.ID)
		if exists {
			continue
		}
		newNodes++
		s.AddNode(newNode)
	}
	s.OrderSet()
	s.resetPointer()
	logrus.WithFields(logrus.Fields{
		"total-nodes-from-db": len(nSet),
		"new-nodes-from-db":   newNodes,
		"total-nodes-in-set":  s.Len(),
	}).Info("updating node-set from db-non-deprecated-set")
}

func (s *NodeOrderedSet) IsPeerAlready(nodeID enode.ID) bool {
	// IsPeerAlready checks whether a peer is already in the Queue.
	s.m.RLock()
	defer s.m.RUnlock()
	_, ok := s.nodeMap[nodeID.String()]
	return ok
}

func (s *NodeOrderedSet) AddNode(hInfo models.HostInfo) {
	logrus.WithField("nodeID", hInfo.ID.String()).Trace("adding node to node-set")
	qNode := NewQueuedNode(hInfo)
	s.m.Lock()
	defer s.m.Unlock()
	s.nodeMap[hInfo.ID.String()] = qNode
	s.nodeList = append([]*QueuedNode{qNode}, s.nodeList[:]...) // add it at the beginning of the queue
}

func (s *NodeOrderedSet) RemoveNode(nodeID enode.ID) {
	logrus.WithField("nodeID", nodeID.String()).Trace("removing node from node-set")
	exists := s.IsPeerAlready(nodeID)
	if !exists {
		logrus.Warn("trying to remove a peer that was no present in the node list")
		return
	}
	s.m.Lock()
	defer s.m.Unlock()
	// from the map
	delete(s.nodeMap, nodeID.String())
	// from the list
	idx := -1
	for i, val := range s.nodeList {
		if val.hostInfo.ID == nodeID {
			idx = i
			break
		}
	}
	// double-check the index
	if idx > -1 {
		s.nodeList = append(s.nodeList[:idx], s.nodeList[idx+1:]...)
	} else {
		logrus.Warn("couldn't find the index of the Node to remove from the NodeList")
	}
}

func (s *NodeOrderedSet) UpdateNodeFromConnAttempt(
	nodeID enode.ID, connAttempt *models.ConnectionAttempt, sameNetwork bool) {
	logEntry := logrus.WithFields(logrus.Fields{
		"nodeID":  nodeID.String(),
		"attempt": connAttempt.Status.String(),
		"error":   connAttempt.Error,
	})
	logEntry.Trace("tracking connection-attempt for node")
	node, exists := s.GetNode(nodeID)
	if !exists {
		logEntry.Warn("connection attempt to a node that is untracked")
	}
	// check the state of the conn attempt
	switch connAttempt.Status {
	case models.SuccessfulConnection:
		// directly remove the peer if th
		if !sameNetwork { // directly prune the node from the list & deprecate it
			connAttempt.Deprecable = true
			s.RemoveNode(nodeID)
			return
		}
		// if possitive, all god
		node.AddPositiveDial(connAttempt.Timestamp)
		connAttempt.Deprecable = false

	case models.FailedConnection:
		// if negative, check if it's deprecable
		connAttempt.Deprecable = node.IsDeprecable()
		if connAttempt.Deprecable {
			s.RemoveNode(nodeID)
		} else {
			node.AddNegativeDial(connAttempt.Timestamp, ParseStateFromError(connAttempt.Error))
		}

	default:
		logrus.WithFields(logrus.Fields{
			"nodeID":  nodeID.String(),
			"attempt": connAttempt.Status.String(),
			"error":   connAttempt.Error,
		}).Warn("unrecognized connection-attempt status for node")
		logrus.Panic("we should have never reached here", connAttempt)
	}
}

// GetNode retrieves the info of the requested node
func (s *NodeOrderedSet) GetNode(nodeID enode.ID) (*QueuedNode, bool) {
	s.m.Lock()
	defer s.m.Unlock()
	p, ok := s.nodeMap[nodeID.String()]
	if !ok {
		return &QueuedNode{}, ok
	}
	return p, ok
}

// IsThereNext returns a boolean indicating whether there is a new item ready to be readed
func (s *NodeOrderedSet) IsThereNext() bool {
	// did we hit the max number of peers to dial?
	if s.nodePtr >= s.Len() {
		return false
	}
	if s.IsEmpty() {
		return false
	}
	return s.nodeList[s.nodePtr].ReadyToDial()
}

func (s *NodeOrderedSet) NextNode() QueuedNode {
	s.m.Lock()
	defer s.m.Unlock()
	if s.nodePtr > s.Len() || s.nodePtr < 0 {
		return QueuedNode{}
	}
	node := s.nodeList[s.nodePtr]
	s.nodePtr++ // this will cause s.IsThereNext() to be negative, handled by the upper layer
	return *node
}

func (s *NodeOrderedSet) resetPointer() {
	logrus.Trace("resetting pointer at NodeSet")
	s.m.Lock()
	defer s.m.Unlock()
	s.nodePtr = 0
}

func (s *NodeOrderedSet) IsEmpty() bool {
	return s.Len() == 0
}

func (s *NodeOrderedSet) Len() int {
	return len(s.nodeList)
}

// ---  SORTING METHODS FOR PeerQueue ----
// OrderSet sorts the items based on their next connection time
func (s *NodeOrderedSet) OrderSet() {
	logrus.Tracef("ordering NodeSet with %d nodes", s.Len())
	s.m.Lock()
	defer s.m.Unlock()
	sort.Sort(s)
}

// Swap is part of sort.Interface.
func (s *NodeOrderedSet) Swap(i, j int) {
	s.nodeList[i], s.nodeList[j] = s.nodeList[j], s.nodeList[i]
}

// Less is part of sort.Interface. We use c.PeerList.NextConnection as the value to sort by.
func (s *NodeOrderedSet) Less(i, j int) bool {
	return s.nodeList[i].NextDialTime().Before(s.nodeList[j].NextDialTime())
}

// --- QueuedNode ---

// Main structure of a node that is queued to be dialed
type QueuedNode struct {
	state           DialState
	hostInfo        models.HostInfo
	nextDialTime    time.Time
	deprecationTime time.Time
}

func NewQueuedNode(hInfo models.HostInfo) *QueuedNode {
	return &QueuedNode{
		state:           ZeroState,
		hostInfo:        hInfo,
		nextDialTime:    time.Time{},
		deprecationTime: time.Time{},
	}
}

func (n *QueuedNode) ReadyToDial() bool {
	return n.nextDialTime.Before(time.Now()) // TODO: should I check if the QueuedNode is empty?
}

func (n *QueuedNode) updateNextDialTime(t time.Time) {
	n.nextDialTime = t
}

func (n *QueuedNode) NextDialTime() time.Time {
	return n.nextDialTime
}

func (n *QueuedNode) IsDeprecable() bool {
	return !n.IsEmpty() && !n.deprecationTime.IsZero() && n.deprecationTime.Before(time.Now())
}

func (n *QueuedNode) IsEmpty() bool {
	return n.hostInfo == models.HostInfo{}
}

func (n *QueuedNode) AddPositiveDial(baseT time.Time) {
	logrus.Trace("adding possitive dial attempt to node", n.hostInfo.ID.String())
	n.state = PossitiveState
	n.nextDialTime = baseT.Add(n.state.DelayFromState())
	n.deprecationTime = time.Time{}
}

func (n *QueuedNode) AddNegativeDial(baseT time.Time, state DialState) {
	logrus.Trace("adding negative dial attempt to node", n.hostInfo.ID.String())
	n.state = state
	n.nextDialTime = baseT.Add(state.DelayFromState())
	if n.deprecationTime.IsZero() {
		n.deprecationTime = baseT.Add(DeprecationMargin)
	}
}
