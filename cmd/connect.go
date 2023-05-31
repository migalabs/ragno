package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/cortze/ragno/crawler"
	"github.com/cortze/ragno/crawler/db"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var RWDeadline time.Duration = 20 * time.Second // for the read and write operations with the remote remoteNodes

var connectOptions struct {
	lvl  string
	enr  string
	file string
}

var ConnectCmd = &cli.Command{
	Name:   "connect",
	Usage:  "connect and identify any given ENR",
	Action: connect,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "log-level",
			Aliases:     []string{"v"},
			Usage:       "sets the verbosity of the logs",
			Value:       "info",
			EnvVars:     []string{"RAGNO_LOG_LEVEL"},
			Destination: &connectOptions.lvl,
		},
		&cli.StringFlag{
			Name:        "enr",
			Usage:       "Enr of the node to connect",
			Aliases:     []string{"e"},
			Required:    false,
			Destination: &connectOptions.enr,
		},
		&cli.StringFlag{
			Name:        "file",
			Usage:       "Path to the csv file with the Enr records to connect",
			Aliases:     []string{"f"},
			Required:    false,
			Destination: &connectOptions.file,
		},
	},
}

func connect(ctx *cli.Context) error {
	logrus.SetLevel(crawler.ParseLogLevel(connectOptions.lvl))

	err := godotenv.Load("../.env")
	if err != nil {
		logrus.Error(".env file couldn't be loaded...")
		panic(err)
	}

	db_name := os.Getenv("POSTGRES_DB")
	user_name := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	postrgres_host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")

	conn_str := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user_name, password, postrgres_host, port, db_name)

	// persister and batchsize ????
	db_manager, err := db.New(ctx.Context, conn_str, 1, 5)
	if err != nil {
		logrus.Error("Couldn't init DB")
		return err
	}

	host, err := crawler.NewHost(
		ctx.Context,
		"0.0.0.0",
		9045,
		// default configuration so far
	)
	if err != nil {
		logrus.Error("failed to create host:")
		return err
	}

	if connectOptions.enr == "" && connectOptions.file == "" {
		logrus.Warn("no Enr or File was provided")
		fmt.Println(connectOptions)
		return nil
	}

	// read the number of enrs
	connectPeers := make([]*enode.Node, 0)
	connectPeers_info := make([][]string, 0)

	if connectOptions.enr != "" {
		rEnr := crawler.ParseStringToEnr(connectOptions.enr)
		connectPeers = append(connectPeers, rEnr)

		currentTime := time.Now()
		// Format the time according to the desired layout
		layout := "2006-01-02 15:04:05.000000 -0700 MST m=+1500.000000000"
		timeString := currentTime.Format(layout)
		info := []string{timeString, timeString, connectOptions.enr}

		connectPeers_info = append(connectPeers_info, info)
	}

	// read the enrs from the given csv file
	if connectOptions.file != "" {
		csvImporter, err := crawler.NewCsvImporter(connectOptions.file)
		if err != nil {
			logrus.Error(err)
			goto connecter
		}
		enrs := csvImporter.Items()
		info := csvImporter.Infos()
		for i, e := range enrs {
			connectPeers = append(connectPeers, e)
			connectPeers_info = append(connectPeers_info, info[i])
		}
	}

connecter:
	// connect and identify the peer
	logrus.Infof("attempting to connect %d nodes", len(connectPeers))
	for i, remoteNode := range connectPeers {
		logrus.Info("connecting to: ", remoteNode)
		hinfo := host.Connect(remoteNode)
		if hinfo.Error != nil {
			logrus.Error(hinfo.Error)
			logrus.Error(`couldn't connect to:`, remoteNode.String())
		}
		// logrus.Infof("remoteNode %s successfully connected:", remoteNode.String())
		// fmt.Println("ID:", remoteNode.ID().String())
		// fmt.Println("PK:", crawler.PubkeyToString(remoteNode.Pubkey()))
		// fmt.Println("Seq:", remoteNode.Seq())
		// fmt.Println("IP:", remoteNode.IP())
		// fmt.Println("TCP:", remoteNode.TCP())
		// fmt.Println("Client:", hinfo.ClientName)
		// fmt.Println("Capabilities:", hinfo.Capabilities)
		// fmt.Println("SoftwareInfo:", hinfo.SoftwareInfo)
		// fmt.Println("Error:", hinfo.Error)

		pubKey := crawler.PubkeyToString(remoteNode.Pubkey())
		err := db_manager.InsertElNode(remoteNode, connectPeers_info[i], hinfo, pubKey)
		if err != nil {
			logrus.Error(err)
		} else {
			logrus.Info("Node succefully saved in DB")
		}
	}
	return nil
}
