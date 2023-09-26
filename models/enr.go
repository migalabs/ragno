package models

import (
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	ogEnr "github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/pkg/errors"
)

type ENRoption func(*ENR) error
type DiscoveryType int8

const (
	UnknownDiscovery DiscoveryType = iota
	Discovery4
	Discovery5
	CsvFile
)

// Basic structure that can be readed from the discovery services
type ENR struct {
	Timestamp time.Time
	Node      *enode.Node
	Record    *ogEnr.Record
	DiscType  DiscoveryType
	ID        enode.ID
	IP        string
	UDP       int
	TCP       int
	Pubkey    string
	Score     int
	Seq       uint64
}

func NewENR(opts ...ENRoption) (*ENR, error) {
	enr := new(ENR)
	for _, opt := range opts {
		err := opt(enr)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create new eth node")
		}
	}
	return enr, nil
}

// WithTimestamp adds any given timestamp into any given ENR
func WithTimestamp(t time.Time) ENRoption {
	return func(enr *ENR) error {
		enr.Timestamp = time.Now()
		return nil
	}
}

// the Discv4 Node won't search for the Eth2data and so on (could be still worth look for it?)
func FromDiscv4(en *enode.Node) ENRoption {
	return func(enr *ENR) error {
		err := en.ValidateComplete()
		if err != nil {
			return err
		}
		// Set First and Last time we saw the Node
		// (if updated, we will only update the LastTime seen)
		enr.Node = en
		enr.Record = en.Record()
		enr.ID = en.ID()
		enr.IP = en.IP().String()
		enr.UDP = en.UDP()
		enr.TCP = en.TCP()
		enr.Seq = en.Seq()
		enr.Pubkey = PubkeyToString(en.Pubkey())
		return nil
	}
}

func FromCSVline(line []string) ENRoption {
	return func(enr *ENR) error {
		node := ParseStringToEnode(line[7]) // Record field
		err := node.ValidateComplete()
		if err != nil {
			return err
		}

		lastSeen, err := time.Parse(time.RFC3339Nano, line[1])
		if err != nil {
			return err
		}
		// apply the readed values
		enr.Timestamp = lastSeen
		enr.Node = node
		enr.Record = node.Record()
		enr.ID = node.ID()
		enr.Pubkey = PubkeyToString(node.Pubkey())
		enr.IP = node.IP().String()
		enr.UDP = node.UDP()
		enr.TCP = node.TCP()
		enr.Seq = node.Seq()
		return nil
	}
}

func (n *ENR) IsValid() bool {
	return (len(n.ID) > 0) && (len(n.IP) > 0) && (n.UDP > 0)
}

func (n ENR) CSVheaders() []string {
	return []string{
		"node_id", "last_seen",
		"ip", "tcp", "udp",
		"seq", "pubkey", "record",
	}
}

func (n *ENR) ComposeCSVItems() []interface{} {
	items := make([]interface{}, 0, 9)
	items = append(items, n.ID.String())
	items = append(items, n.Timestamp.String())
	items = append(items, n.IP)
	items = append(items, strconv.Itoa(n.TCP))
	items = append(items, strconv.Itoa(n.UDP))
	items = append(items, strconv.Itoa(int(n.Seq)))
	items = append(items, n.Pubkey)
	items = append(items, n.Node.String())
	return items
}
