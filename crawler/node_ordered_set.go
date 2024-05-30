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

// Recieves a list of HostInfo and adds the nodes to the set (if they are not yet added). O(m+n*m) m being the amount of elements to try to insert and n the length of the set. Adding many new nodes can be expensive.
func (s *NodeOrderedSet) UpdateSetFromList(newNodeList []models.HostInfo) {
	// try to add the missing nodes from the DB into the NodeSet
	prevLen := s.Len()
	for _, newNode := range newNodeList {
		s.addNode(newNode)
	}
	s.orderSet()
	s.resetPointer()
	logrus.WithFields(logrus.Fields{
		"total-nodes-from-db": len(newNodeList),
		"new-nodes-from-db":   s.Len() - prevLen,
		"total-nodes-in-set":  s.Len(),
	}).Info("updating node-set from db-non-deprecated-set")
}

// Checks if peer node is already inserted. O(1)
func (s *NodeOrderedSet) IsPeerAlready(nodeID enode.ID) bool {
	// IsPeerAlready checks whether a peer is already in the Queue.
	s.m.RLock()
	defer s.m.RUnlock()
	_, ok := s.nodeMap[nodeID.String()]
	return ok
}

// Adds a new node to the set. O(n)
func (s *NodeOrderedSet) addNode(hInfo models.HostInfo) {
	logrus.WithField("nodeID", hInfo.ID.String()).Trace("adding node to node-set")
	exists := s.IsPeerAlready(hInfo.ID)
	if exists {
		return
	}
	qNode := NewQueuedNode(hInfo)
	s.m.Lock()
	defer s.m.Unlock()
	s.nodeMap[hInfo.ID.String()] = qNode
	s.nodeList = append([]*QueuedNode{qNode}, s.nodeList[:]...) // add it at the beginning of the queue
}

// Removes a node from the ordered set by nodeID. If the node isn't in the set, does nothing. O(n)
func (s *NodeOrderedSet) RemoveNode(nodeID enode.ID) {
	logrus.WithField("nodeID", nodeID.String()).Trace("removing node from node-set")
	exists := s.IsPeerAlready(nodeID)
	if !exists {
		logrus.Warn("trying to remove a peer that was not present in the node list")
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

// GetNode retrieves a pointer to a node in the set. If the node isn't in the set, it returns an empty node. O(1)
func (s *NodeOrderedSet) GetNode(nodeID enode.ID) (*QueuedNode, bool) {
	s.m.Lock()
	defer s.m.Unlock()
	p, ok := s.nodeMap[nodeID.String()]
	if !ok {
		return &QueuedNode{}, ok
	}
	return p, ok
}

// IsThereNext returns a boolean indicating whether there is a new item ready to be read. Used for iterating. O(1)
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

// NextNode returns the next node in the ordered set.
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

// Resets the pointer used for getting the NextNode. O(1)
func (s *NodeOrderedSet) resetPointer() {
	logrus.Trace("resetting pointer at NodeSet")
	s.m.Lock()
	defer s.m.Unlock()
	s.nodePtr = 0
}

// Checks if there are no nodes in the set. O(1)
func (s *NodeOrderedSet) IsEmpty() bool {
	return s.Len() == 0
}

// Returns the length of the set. O(1)
func (s *NodeOrderedSet) Len() int {
	return len(s.nodeList)
}

// ---  SORTING METHODS FOR PeerQueue ----

// OrderSet sorts the items based on their next connection time. O(n*log(n))
func (s *NodeOrderedSet) orderSet() {
	logrus.Tracef("ordering NodeSet with %d nodes", s.Len())
	s.m.Lock()
	defer s.m.Unlock()
	sort.Sort(s)
}

// Swap is part of sort.Interface. Not intended to be called manually
func (s *NodeOrderedSet) Swap(i, j int) {
	s.nodeList[i], s.nodeList[j] = s.nodeList[j], s.nodeList[i]
}

// Less is part of sort.Interface. We use c.PeerList.NextConnection as the value to sort by. Not intended to be called manually
func (s *NodeOrderedSet) Less(i, j int) bool {
	return s.nodeList[i].NextDialTime().Before(s.nodeList[j].NextDialTime())
}
