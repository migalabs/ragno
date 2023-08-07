package modules

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
)

type EnodeSet struct {
	m    sync.RWMutex
	list map[string]*EthNode
}

func NewEnodeSet() *EnodeSet {
	return &EnodeSet{
		list: make(map[string]*EthNode),
	}
}

func (s *EnodeSet) AddNode(n *EthNode) error {
	if !n.IsValid() {
		return errors.New(fmt.Sprintf("attempt to persist non-valid node %+v", n))
	}
	s.m.Lock()
	defer s.m.Unlock()
	oldnode, ok := s.list[n.Node.String()]
	if ok {
		oldnode.Update(n)
		s.list[n.ID] = oldnode
	} else {
		s.list[n.ID] = n
	}
	return nil
}

func (s *EnodeSet) PeerList() []*EthNode {
	plist := make([]*EthNode, 0, len(s.list))
	s.m.Lock()
	defer s.m.Unlock()
	for _, v := range s.list {
		plist = append(plist, v)
	}
	return plist
}

func (s *EnodeSet) Len() int {
	s.m.RLock()
	defer s.m.RUnlock()
	return len(s.list)
}

// In relation to the Ethereum Node

// TODO: so far only contemplated the discv4 version of an eth node
type EthNode struct {
	Node   *enode.Node
	FirstSeen time.Time
	LastSeen  time.Time
	ID     string
	IP     string
	UDP    int
	TCP    int
	Pubkey string
	Score  int
	Seq    uint64
	Record *enr.Record
}

type EthNodeOption func(*EthNode) error

func NewEthNode(opts ...EthNodeOption) (*EthNode, error) {
	ethNode := new(EthNode)
	for _, opt := range opts {
		err := opt(ethNode)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create new eth node")
		}
	}
	return ethNode, nil
}

// the Discv4 Node won't search for the Eth2data and so on (could be still worth look for it?)
func FromDiscv4Node(en *enode.Node) EthNodeOption {
	return func(ethNode *EthNode) error {
		err := en.ValidateComplete()
		if err != nil {
			return err
		}
		// Set First and Last time we saw the Node
		// (if updated, we will only update the LastTime seen)
		ethNode.FirstSeen = time.Now()
		ethNode.LastSeen = time.Now()
		ethNode.Node = en
		ethNode.ID = en.ID().String()
		ethNode.UDP = en.UDP()
		ethNode.TCP = en.TCP()
		ethNode.IP = en.IP().String()
		ethNode.Seq = en.Seq()
		ethNode.Record = en.Record()
		ethNode.Pubkey = PubkeyToString(en.Pubkey())
		return nil
	}
}

func (n *EthNode) Update(en *EthNode) {
	// Only update the LastTime seen
	n.LastSeen = en.LastSeen
	n.Node = en.Node
	n.ID = en.ID
	n.UDP = en.UDP
	n.TCP = en.TCP
	n.IP = en.IP
	n.Seq = en.Seq
	n.Record = en.Record
	n.Pubkey = en.Pubkey
}

func (n *EthNode) IsValid() bool {
	return (len(n.ID) > 0) && (len(n.IP) > 0) && (n.UDP > 0)
}

func (n *EthNode) ComposeCSVItems() []string {
	items := make([]string, 0, 9)
	items = append(items, n.ID)
	items = append(items, n.FirstSeen.String())
	items = append(items, n.LastSeen.String())
	items = append(items, n.IP)
	items = append(items, strconv.Itoa(n.TCP))
	items = append(items, strconv.Itoa(n.UDP))
	items = append(items, strconv.Itoa(int(n.Seq)))
	items = append(items, n.Pubkey)
	items = append(items, n.Node.String())
	return items
}
