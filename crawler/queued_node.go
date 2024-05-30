package crawler

import (
	"time"

	"github.com/cortze/ragno/models"
	"github.com/sirupsen/logrus"
)

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
