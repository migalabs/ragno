package peerDiscoverer

import (
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/cortze/ragno/modules"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/params"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type Discv4PeerDiscoverer struct {
	port int
}

func NewDisv4PeerDiscoverer(port int) (PeerDiscoverer, error) {
	logrus.Info("Using Discv4 peer discoverer")

	disc := &Discv4PeerDiscoverer{
		port: port,
	}
	return disc, nil
}

func (d *Discv4PeerDiscoverer) Run(sendingChan chan *modules.ELNode) error {
	// create a new context
	ctx := cli.NewContext(nil, nil, nil)

	var wg sync.WaitGroup
	doneC := make(chan struct{}, 1)

	wg.Add(1)
	err := d.runDiscv4Service(ctx, &wg, doneC, sendingChan)
	if err != nil {
		return errors.Wrap(err, "unable to run the Discv4 service")
	}

	closeC := make(chan os.Signal, 1)
	signal.Notify(closeC, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	// Make it pretty
	select {
	case <-ctx.Done():
		logrus.Error("Context died")
		return nil

	case <-closeC:
		logrus.Info("Shutdown detected")
		doneC <- struct{}{}
		wg.Wait()
		return nil
	}
}

func (d *Discv4PeerDiscoverer) sendNodes(sendingChan chan *modules.ELNode, node *modules.ELNode) {
	sendingChan <- node
}

func (d *Discv4PeerDiscoverer) runDiscv4Service(ctx *cli.Context, wg *sync.WaitGroup, doneC chan struct{}, sendC chan *modules.ELNode) error {
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
				ethNode, err := modules.NewEthNode(
					modules.FromDiscv4Node(node),
				)
				if err != nil {
					logrus.Error(errors.Wrap(err, "unable to add new node"))
				}
				elNode := modules.ELNode{
					Enode:         ethNode.Node,
					Enr:           ethNode.Node.String(),
					FirstTimeSeen: ethNode.FirstT.String(),
					LastTimeSeen:  ethNode.LastT.String(),
				}
				d.sendNodes(sendC, &elNode)
			}
		}
	}()
	return nil
}
