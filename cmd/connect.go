package cmd

import (
	"fmt"
	"time"

	"github.com/cortze/ragno/crawler"

	"github.com/ethereum/go-ethereum/p2p/enode"
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
	
	host, err := crawler.NewHost(
		ctx.Context,
		"0.0.0.0",
		9045,
		// default configuration so far
	) 
	if err != nil {
		logrus.Error("failed to create host:",)
		return err
	}

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
	// connect and identify the peer
	logrus.Infof("attempting to connect %d nodes", len(connectPeers))
	for _, remoteNode := range connectPeers {
		logrus.Info("connecting to node", remoteNode)
		hinfo := host.Connect(remoteNode)
		if hinfo.Error != nil {
			logrus.Error("failed to connect %s", remoteNode.String())
			continue
		}
		logrus.Infof("remoteNode %s successfully connected:", remoteNode.String())
		fmt.Println("ID:", remoteNode.ID().String())
		fmt.Println("IP:", crawler.PubkeyToString(remoteNode.Pubkey()))
		fmt.Println("Seq:", remoteNode.Seq())
		fmt.Println("IP:", remoteNode.IP())
		fmt.Println("TCP:", remoteNode.TCP())
		fmt.Println("Client:", hinfo.ClientName)
		fmt.Println("Capabilities:", hinfo.Capabilities)
		fmt.Println("SoftwareInfo:", hinfo.SoftwareInfo)
		fmt.Println("Error:", hinfo.Error)
	}
	return nil
}

