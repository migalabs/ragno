package crawler

import (
	"testing"

	"github.com/cortze/ragno/models"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

func TestNodeOrderedSetAddRemove(t *testing.T) {
	nodes := NewNodeOrderedSet()

	if nodes.Len() > 0 || !nodes.IsEmpty() {
		t.Error("Expected node set to be empty")
	}

	nodeId1 := enode.HexID("1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	hostInfo1 := models.HostInfo{
		ID: nodeId1,
	}

	nodes.addNode(hostInfo1)
	if !nodes.IsPeerAlready(nodeId1) {
		t.Error("Expected node1 to be on node set")
	}
	if nodes.IsEmpty() {
		t.Error("Node set should not be empty")
	}
	if nodes.Len() != 1 {
		t.Errorf("Node set lenght is %d instead of the expected 1", nodes.Len())
	}
	nodes.addNode(hostInfo1)
	if nodes.Len() != 1 {
		t.Errorf("Node set lenght is %d instead of the expected 1", nodes.Len())
	}

	nodeId2 := enode.HexID("2aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	hostInfo2 := models.HostInfo{
		ID: nodeId2,
	}
	nodes.addNode(hostInfo2)
	if !nodes.IsPeerAlready(nodeId2) {
		t.Error("Expected node2 to be on node set")
	}
	if nodes.Len() != 2 {
		t.Errorf("Node set lenght is %d instead of the expected 2", nodes.Len())
	}

	nodeId3 := enode.HexID("3aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	hostInfo3 := models.HostInfo{
		ID: nodeId3,
	}
	nodes.addNode(hostInfo3)
	if !nodes.IsPeerAlready(nodeId3) {
		t.Error("Expected node3 to be on node set")
	}
	if nodes.Len() != 3 {
		t.Errorf("Node set lenght is %d instead of the expected 3", nodes.Len())
	}

	nodes.RemoveNode(nodeId1)
	if nodes.IsPeerAlready(nodeId1) {
		t.Error("Expected node1 to not be on node set")
	}
	if nodes.Len() != 2 {
		t.Errorf("Node set lenght is %d instead of the expected 2", nodes.Len())
	}

	nodes.RemoveNode(nodeId3)
	if nodes.IsPeerAlready(nodeId3) {
		t.Error("Expected node3 to not be on node set")
	}
	if nodes.Len() != 1 {
		t.Errorf("Node set lenght is %d instead of the expected 1", nodes.Len())
	}
	nodes.RemoveNode(nodeId2)
	if nodes.IsPeerAlready(nodeId2) {
		t.Error("Expected node2 to not be on node set")
	}
	nodes.RemoveNode(nodeId2)

	if nodes.Len() != 0 || !nodes.IsEmpty() {
		t.Errorf("Node set lenght is %d instead of the expected 0", nodes.Len())
	}
}
