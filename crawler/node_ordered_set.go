package crawler

import (
	"sort"
	"sync"

	"github.com/cortze/ragno/models"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/sirupsen/logrus"
)

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

