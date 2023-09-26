package peerdiscovery

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/cortze/ragno/models"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/params"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type Discv4 struct {
	port      int
	discvType models.DiscoveryType
	enrC      chan *models.ENR
	doneC     chan struct{}
	wg        sync.WaitGroup
}

func NewDiscv4(port int) (*Discv4, error) {
	logrus.Info("Using Discv4 peer discoverer")

	disc := &Discv4{
		port:      port,
		enrC:      make(chan *models.ENR),
		discvType: models.Discovery4,
	}
	return disc, nil
}

func (d *Discv4) Run() (chan *models.ENR, error) {
	// create a new context
	ctx := cli.NewContext(nil, nil, nil)

	d.wg.Add(1)
	return d.runDiscv4Service(ctx, d.doneC)
}

func (d *Discv4) runDiscv4Service(ctx *cli.Context, doneC chan struct{}) (chan *models.ENR, error) {
	var err error

	// create the private Key
	privKey, err := crypto.GenerateKey()
	if err != nil {
		return d.enrC, nil
	}

	bnodes := params.MainnetBootnodes
	bootnodes, err := models.ParseBootnodes(bnodes)
	if err != nil {
		return d.enrC, nil
	}

	ethDB, err := enode.OpenDB("")
	if err != nil {
		return d.enrC, nil
	}

	localNode := enode.NewLocalNode(ethDB, privKey)
	udpAddr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: d.port,
	}
	udpListener, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return d.enrC, nil
	}

	discv4Options := discover.Config{
		PrivateKey: privKey,
		Bootnodes:  bootnodes,
	}
	discoverer4, err := discover.ListenV4(udpListener, localNode, discv4Options)
	if err != nil {
		return d.enrC, nil
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
			discoverer4.Close()
			logrus.Info("discv4 down")
			d.wg.Done()
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
					"pubkey": models.PubkeyToString(node.Pubkey()),
				}).Debug("new discv4 node")
				enr, err := models.NewENR(
					node,
					models.FromDiscv4(node),
					models.WithTimestamp(time.Now()))
				if err != nil {
					logrus.Error(errors.Wrap(err, "unable to add new node"))
				}
				d.notifyNewNode(enr)
			}
		}
	}()
	return d.enrC, nil
}

func (d *Discv4) notifyNewNode(enr *models.ENR) {
	d.enrC <- enr
}

func (d *Discv4) Close() {
	// notify of closure
	d.doneC <- struct{}{}
	d.wg.Wait()
	close(d.doneC)
}
