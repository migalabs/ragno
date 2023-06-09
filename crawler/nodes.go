package crawler

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/cmd/devp2p/tooling/ethtest"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
)

const (
	// csv columns
	NODE_ID = iota
	FIRST_SEEN
	LAST_SEEN
	IP
	TCP
	UDP
	SEQ
	PK
	ENR
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
	FirstT time.Time
	LastT  time.Time
	ID     string
	IP     string
	UDP    int
	TCP    int
	Pubkey string
	Score  int
	Seq    uint64
	Record *enr.Record
}

type ELNodeInfo struct {
	Enode         *enode.Node
	Hinfo         ethtest.HandshakeDetails
	Enr           string
	FirstTimeSeen string
	LastTimeSeen  string
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
		ethNode.FirstT = time.Now()
		ethNode.LastT = time.Now()
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
	n.LastT = en.LastT
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
	items = append(items, n.FirstT.String())
	items = append(items, n.LastT.String())
	items = append(items, n.IP)
	items = append(items, strconv.Itoa(n.TCP))
	items = append(items, strconv.Itoa(n.UDP))
	items = append(items, strconv.Itoa(int(n.Seq)))
	items = append(items, n.Pubkey)
	items = append(items, n.Node.String())
	return items
}

func ParseStringToEnr(enr string) *enode.Node {
	// parse the Enr
	remoteEnr, err := enode.Parse(enode.ValidSchemes, enr)
	if err != nil {
		remoteEnr = enode.MustParseV4(enr)
	}
	return remoteEnr
}

func ParseCsvToNodeInfo(csvImp CSVImporter) ([]*ELNodeInfo, error) {
	// get all the lines from the CSV
	lines, err := csvImp.Items()
	if err != nil {
		return nil, err
	}

	lines = lines[1:]

	// create the list of ELNodeInfo
	enrs := make([]*ELNodeInfo, 0, len(lines)-1)

	// parse the file
	for _, line := range lines {
		// create the ELNodeInfo
		elNodeInfo := new(ELNodeInfo)
		elNodeInfo.Enode = ParseStringToEnr(line[ENR])
		elNodeInfo.Enr = line[ENR]
		elNodeInfo.FirstTimeSeen = line[FIRST_SEEN]
		elNodeInfo.LastTimeSeen = line[LAST_SEEN]
		// add the struct to the list
		enrs = append(enrs, elNodeInfo)
	}
	return enrs, nil
}
