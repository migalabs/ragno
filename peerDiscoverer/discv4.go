package peerDiscoverer

import (
	"github.com/cortze/ragno/modules"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/params"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"net"
	"strconv"
	"sync"
)

type Discv4 struct {
	port  int
	doneC chan struct{}
}

func NewDiscv4(port int) (*Discv4, error) {
	logrus.Info("Using Discv4 peer discoverer")

	disc := &Discv4{
		port: port,
	}
	return disc, nil
}

func (d *Discv4) Run(sendingChan chan *modules.ELNode) error {
	// create a new context
	ctx := cli.NewContext(nil, nil, nil)

	var wg sync.WaitGroup
	d.doneC = make(chan struct{}, 1)

	wg.Add(1)
	return d.runDiscv4Service(ctx, &wg, d.doneC, sendingChan)
}

func (d *Discv4) runDiscv4Service(ctx *cli.Context, wg *sync.WaitGroup, doneC chan struct{}, sendC chan *modules.ELNode) error {
	var err error

	// create the private Key
	privKey, err := crypto.GenerateKey()
	if err != nil {
		return err
	}

	bnodes := params.MainnetBootnodes
	bootnodes, err := modules.ParseBootnodes(bnodes)
	if err != nil {
		return err
	}

	ethDB, err := enode.OpenDB("")
	if err != nil {
		return err
	}

	localNode := enode.NewLocalNode(ethDB, privKey)
	udpAddr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: d.port,
	}
	udpListener, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	discv4Options := discover.Config{
		PrivateKey: privKey,
		Bootnodes:  bootnodes,
	}
	discoverer4, err := discover.ListenV4(udpListener, localNode, discv4Options)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"ip":   udpAddr.IP.String(),
		"port": strconv.Itoa(udpAddr.Port),
	}).Info("launching rango discv4")

	closeC := make(chan struct{})
	// actuall loop for crawling
	go func() {
		// finish the wg
		defer func() {
			wg.Done()
			logrus.Info("discv4 down")
			closeC <- struct{}{}
		}()
		// generate an iterator
		rNodes := discoverer4.RandomNodes()
		defer rNodes.Close()
		for rNodes.Next() {
			select {
			case <-ctx.Context.Done():
				logrus.Error("unhandled ctx received")
				return
			case <-doneC:
				logrus.Info("Shutdown detected")
				return
			default:
				// if everthing okey and no errors raised, discover more nodes
				node := rNodes.Node()
				logrus.WithFields(logrus.Fields{
					"enr":    node.String(),
					"ID":     node.ID(),
					"IP":     node.IP(),
					"UDP":    node.UDP(),
					"TCP":    node.TCP(),
					"seq":    node.Seq(),
					"pubkey": modules.PubkeyToString(node.Pubkey()),
				}).Debug("new discv4 node")
				ethNode, err := modules.NewEthNode(modules.FromDiscv4Node(node))
				if err != nil {
					logrus.Error(errors.Wrap(err, "unable to add new node"))
				}
				elNode := modules.ELNode{
					Enode:         ethNode.Node,
					Enr:           ethNode.Node.String(),
					FirstTimeSeen: ethNode.FirstSeen,
					LastTimeSeen:  ethNode.LastSeen,
				}
				d.SendNodes(sendC, &elNode)
			}
		}
	}()
	return nil
}

func (d *Discv4) SendNodes(sendingChan chan *modules.ELNode, node *modules.ELNode) {
	sendingChan <- node
}

func (d *Discv4) Close() error {
	// notify of closure
	d.doneC <- struct{}{}
	close(d.doneC)
	return nil
}
