package cmd

import (
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"strconv"

	"github.com/cortze/ragno/crawler"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/params"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)


var discv4Configuration struct {
	logLevel string
	outputFile string
	port int
}


var Discv4Cmd = &cli.Command{
	Name: "discv4",
	Usage: "discover4 prints nodes in the discovery4 network",
	Action: discover4, 
	Flags: []cli.Flag {
		&cli.StringFlag{
			Name: "log-level",
			Aliases: []string{"v"},
			Usage: "sets the verbosity of the logs",
			Value: "info",
			EnvVars: []string{"RAGNO_LOG_LEVEL"},
			Destination: &discv4Configuration.logLevel,
		}, 
		&cli.StringFlag{
			Name: "output",
			Aliases: []string{"f"},
			Usage: "path to the file where the output of the file will be flushed",
			Value: "ragno_crawl.csv",
			EnvVars: []string{"RAGNO_OUTPUT"},
			Destination: &discv4Configuration.outputFile,
		},
		&cli.IntFlag{
			Name: "port",
			Aliases: []string{"p"},
			Usage: "port that the tool will use for discovery purposes",
			Value: 9045,
			EnvVars: []string{"RAGNO_PORT"},
			Destination: &discv4Configuration.port,
		},

	},
}

func discover4(ctx *cli.Context) error {
	// set log level
	logrus.SetLevel(crawler.ParseLogLevel(discv4Configuration.logLevel))

	// control variables
	var wg sync.WaitGroup
	doneC := make(chan struct{}, 1)
	
	wg.Add(1)
	err := runDiscv4Service(ctx, &wg, doneC, discv4Configuration.port, discv4Configuration.outputFile)
	if err != nil {
		return errors.Wrap(err, "unable to run the Discv4 service")
	}

	closeC := make(chan os.Signal, 1)
	signal.Notify(closeC, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	// Make it pretty 
	select {
		case <- ctx.Context.Done():
			logrus.Error("Context died")
			return nil
			
		case <- closeC:
			logrus.Info("Shutdown detected")
			doneC <- struct{}{}
			wg.Wait()
			return nil
	}
}


func runDiscv4Service(ctx *cli.Context, wg *sync.WaitGroup, doneC chan struct{}, port int, output string) error {
	var err error
	// create the private Key
	privKey, err := crypto.GenerateKey()
	if err != nil {
		return err
	}
	bnodes := params.MainnetBootnodes
	bootnodes, err := crawler.ParseBootnodes(bnodes)
	if err != nil {
		return err
	}
	ethDB, err := enode.OpenDB("")
	if err != nil {
		return err
	}
	localNode := enode.NewLocalNode(ethDB, privKey)

	udpAddr := &net.UDPAddr{
		IP: net.IPv4zero,
		Port: port,
	}
	udpListener, err := net.ListenUDP("udp", udpAddr)
	if err  != nil {
		return err
	}

	discv4Options := discover.Config{
		PrivateKey: privKey,
		Bootnodes: bootnodes,
	}
	discoverer4, err := discover.ListenV4(udpListener, localNode, discv4Options)
	if err != nil {
		return err
	}

	var headers []string = []string{
		"node_id", "first_seen", "last_seen",
		"ip", "tcp", "udp", 
		"seq", "pubkey", "record",
	}
	csvExp, err := crawler.NewCsvExporter(output, headers)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"log-level": discv4Configuration.logLevel,
		"ip": udpAddr.IP.String(),
		"port": strconv.Itoa(udpAddr.Port),
		"output-file": discv4Configuration.outputFile,
	}).Info("launching rango discv4")
	// compose the nodeset
	nodeSet := crawler.NewEnodeSet()
	closeC := make(chan struct{})
	// actuall loop for crawling
	go func() {
		// finish the wg 
		defer func (){ 
			wg.Done()
			logrus.Info("discv4 down")
			closeC <- struct{}{}
		}()
		// generate an iterator	
		rNodes := discoverer4.RandomNodes()	
		defer rNodes.Close()
		for rNodes.Next() {
			select {
			case <- ctx.Context.Done():
				logrus.Error("unhandled ctx received")
				return 
			case <- doneC:
				logrus.Info("Shutdown detected")
				return
			default:
				// if everthing okey and no errors raised, discover more nodes
				node := rNodes.Node() 	
				logrus.WithFields(logrus.Fields{
					"enr": node.String(),
					"ID": node.ID(),
					"IP": node.IP(),
					"UDP": node.UDP(),
					"TCP": node.TCP(),
					"seq": node.Seq(),
					"pubkey": crawler.PubkeyToString(node.Pubkey()),
				}).Debug("new discv4 node")
				ethNode, err := crawler.NewEthNode(
					crawler.FromDiscv4Node(node),
				)
				if err != nil {
					logrus.Error(errors.Wrap(err, "unable to add new node"))
				}
				err = nodeSet.AddNode(ethNode)
				if err != nil {
					logrus.Error(errors.Wrap(err, "unable to store new node"))
				}
			}
		}
	}()

	// persister routine 
	wg.Add(1)
	go func() {
		defer func (){ 
			wg.Done()
			csvExp.Close()
			logrus.Info("discv4 persister down")
		}()
		persistT := time.NewTicker(10*time.Second)	
		for {
			select {
			case <- closeC:
				logrus.Info("shutdown detected (persister)")	
				csvExp.Export(nodeSet.PeerList())
				return
			case <- persistT.C:
				logrus.Infof("flushing peer list (%d peers) to csv", nodeSet.Len())
				csvExp.Export(nodeSet.PeerList())
				persistT.Reset(10*time.Second)
			}
		}
	}()
	return nil
}

