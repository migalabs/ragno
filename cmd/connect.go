package cmd

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/cortze/ragno/crawler"
	"github.com/ethereum/go-ethereum/cmd/devp2p/tooling/ethtest" 
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/forkid"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/rlpx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var RWDeadline time.Duration= 20 * time.Second // for the read and write operations with the remote remoteNodes

var connectOptions struct {
	lvl string
	enr string
	file string
}

var ConnectCmd = &cli.Command{
	Name: "connect",
	Usage: "connect and identify any given ENR",
	Action: connect, 
	Flags: []cli.Flag {
		&cli.StringFlag{
			Name: "log-level",
			Aliases: []string{"v"},
			Usage: "sets the verbosity of the logs",
			Value: "info",
			EnvVars: []string{"RAGNO_LOG_LEVEL"},
			Destination: &connectOptions.lvl,
		},
		&cli.StringFlag{
			Name: "enr",
			Usage: "Enr of the node to connect",
			Aliases: []string{"e"},
			Required: false,
			Destination: &connectOptions.enr,
		},
		&cli.StringFlag{
			Name: "file",
			Usage: "Path to the csv file with the Enr records to connect",
			Aliases: []string{"f"},
			Required: false,
			Destination: &connectOptions.file,
		},
	},
}

func connect(ctx *cli.Context) error {
	logrus.SetLevel(crawler.ParseLogLevel(connectOptions.lvl))
	if connectOptions.enr == "" && connectOptions.file == "" {
		logrus.Warn("no Enr or File was provided")
		fmt.Println(connectOptions)
		return nil
	}
	// read the number of enrs 
	connectPeers := make([]*enode.Node, 0)
	if connectOptions.enr != "" {
		rEnr := crawler.ParseStringToEnr(connectOptions.enr)
		connectPeers = append(connectPeers, rEnr)
	}
	
	// read the enrs from the given csv file
	if connectOptions.file != "" {
		csvImporter, err := crawler.NewCsvImporter(connectOptions.file)
		if err != nil  {
			logrus.Error(err)
			goto connecter	
		}
		enrs := csvImporter.Items()
		for _, e := range enrs {
			connectPeers = append(connectPeers, e)
		}
	}
	
	connecter:	
	// connect the node (whole sequence)
	privKey, _ := crypto.GenerateKey()
	
	logrus.Infof("attempting to connect %d nodes", len(connectPeers))
	for _, remoteNode := range connectPeers {
		logrus.Info("connecting to node", remoteNode)
		hinfo := connectNode(remoteNode, privKey)
		if hinfo.Error != nil {
			logrus.Error("failed to connect %s", remoteNode.String())
			continue
		}

		logrus.Infof("remoteNode %s successfully connected:", remoteNode.String())
		fmt.Println("ID:", remoteNode.ID().String())
		fmt.Println("IP:", remoteNode.IP())
		fmt.Println("TCP:", remoteNode.TCP())
		fmt.Println("Client:", hinfo.ClientName)
		fmt.Println("Capabilities:", hinfo.Capabilities)
		fmt.Println("SoftwareInfo:", hinfo.SoftwareInfo)
		fmt.Println("Error:", hinfo.Error)
	}
	return nil
}


func connectNode(remoteN *enode.Node, priv *ecdsa.PrivateKey) ethtest.HandshakeDetails {
	conn, details := dial(remoteN, priv)
	if details.Error != nil {
		return details
	}
	defer conn.Close()
	return details
}

func dial(n *enode.Node, priv *ecdsa.PrivateKey) (ethtest.Conn, ethtest.HandshakeDetails) {
	netConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", n.IP(),n.TCP())); 
	if err != nil {
		return ethtest.Conn{}, ethtest.HandshakeDetails{Error: errors.Wrap(err, "unable to net.dial node")}
	}
	conn:= ethtest.Conn{
		Conn: rlpx.NewConn(netConn, n.Pubkey()),
	}

	_, err = conn.Handshake(priv)
	if err != nil {
		return ethtest.Conn{}, ethtest.HandshakeDetails{Error: err} 
	}


	details := makeHelloHandshake(&conn, priv)	
	if details.Error != nil {
		conn.Close()
		return conn, ethtest.HandshakeDetails{Error: errors.Wrap(err, "unable to initiate Handshake with node")}
	}
	return conn, details
}

func makeHelloHandshake(conn *ethtest.Conn, priv *ecdsa.PrivateKey) ethtest.HandshakeDetails {
	ourCaps := []p2p.Cap{
		{Name: "eth", Version: 66},
		{Name: "eth", Version: 67},
		{Name: "eth", Version: 68},
	}
	highestProtoVersion := uint(68)
	return conn.DetailedHandshake(priv, ourCaps, highestProtoVersion)
}


type HostInfo struct {
	IP string 
	Port int 
	ClientType string 
	NetworkID uint64
	Capabilities []p2p.Cap
	ForkID forkid.ID
	Blockheight     string
	TotalDifficulty *big.Int
	HeadHash        common.Hash
}

