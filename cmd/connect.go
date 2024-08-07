package cmd

import (
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/cortze/ragno/crawler"
	"github.com/cortze/ragno/models"
)

var RWDeadline time.Duration = 20 * time.Second // for the read and write operations with the remote remoteNodes

var (
	DefaultHostIP   = "0.0.0.0"
	DefaultHostPort = 9050
	DefaultLogLevel = "info"
)

var connectOptions struct {
	lvl          string
	enr          string
	hostIP       string
	hostPort     int
	connTimeout  time.Duration
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
			Name:        "host-ip",
			Usage:       "IP address of the host",
			Aliases:     []string{"i"},
			Destination: &connectOptions.hostIP,
		},
		&cli.IntFlag{
			Name:        "host-port",
			Usage:       "Port of the host",
			Aliases:     []string{"p"},
			Destination: &connectOptions.hostPort,
		},
		&cli.StringFlag{
			Name:        "enr",
			Usage:       "Enr of the node to connect",
			Aliases:     []string{"e"},
			Required:    true,
			Destination: &connectOptions.enr,
		},
		&cli.DurationFlag{
			Name:        "conn-timeout",
			Usage:       "Timeout in seconds for connection",
			Aliases:     []string{"ct"},
			Required:    true,
			Destination: &connectOptions.connTimeout,
		},
	},
}

func connect(ctx *cli.Context) error {
	// create a host
	if connectOptions.hostIP == "" {
		connectOptions.hostIP = DefaultHostIP
	}
	if connectOptions.hostPort == 0 {
		connectOptions.hostPort = DefaultHostPort
	}
	host, err := crawler.NewHost(
		ctx.Context,
		connectOptions.hostIP,
		connectOptions.hostPort,
		connectOptions.connTimeout,
	)
	if err != nil {
		logrus.Error("failed to create host:")
		return err
	}

	node := models.ParseStringToEnode(connectOptions.enr)
	enr, err := models.NewENR(models.FromDiscv4(node))
	if err != nil {
		return errors.Wrap(err, "unable to parse ENR")
	}

	details, chain, err := host.Connect(enr.GetHostInfo())
	if err != nil {
		logrus.Info("Couldn't connect to Node: ", enr.ID, ": ", err)
		return nil
	}

	logrus.Info("Connected to Node: ", enr.Node.String())
	logrus.Info("Node's IP: ", enr.IP)
	logrus.Info("Node's TCP: ", enr.TCP)
	logrus.Info("Node's UDP: ", enr.UDP)
	logrus.Info("Node's ID: ", enr.ID.String())
	logrus.Info("Node's Pubkey: ", enr.Pubkey)
	logrus.Info("Node's Seq: ", enr.Seq)
	// node info
	logrus.Info("Node's Client: ", details.ClientName)
	logrus.Info("Node's Capabilities: ", details.Capabilities)
	logrus.Info("Node's SoftwareInfo: ", details.SoftwareInfo)
	// chain details
	logrus.Info("Node's NetworkID: ", chain.NetworkID)
	logrus.Info("Node's ForkID: ", chain.ForkID)
	logrus.Info("Node's ProtocolVersion: ", chain.ProtocolVersion)
	logrus.Info("Node's HeadHash: ", chain.HeadHash)
	logrus.Info("Node's TotalDiff: ", chain.TotalDifficulty)
	return nil
}
