package cmd

import (
	csvs "github.com/cortze/ragno/csv"
	"github.com/cortze/ragno/models"
	"github.com/cortze/ragno/peerdiscovery"
	"github.com/ethereum/go-ethereum/log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/cortze/ragno/crawler"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var discv4Configuration struct {
	logLevel   string
	outputFile string
	port       int
}

var Discv4Cmd = &cli.Command{
	Name:   "discv4",
	Usage:  "discover4 prints nodes in the discovery4 network",
	Action: discover4,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "log-level",
			Aliases:     []string{"v"},
			Usage:       "sets the verbosity of the logs",
			Value:       "info",
			EnvVars:     []string{"RAGNO_LOG_LEVEL"},
			Destination: &discv4Configuration.logLevel,
		},
		&cli.StringFlag{
			Name:        "output",
			Aliases:     []string{"f"},
			Usage:       "path to the file where the output of the file will be flushed",
			Value:       "ragno_crawl.csv",
			EnvVars:     []string{"RAGNO_OUTPUT"},
			Destination: &discv4Configuration.outputFile,
		},
		&cli.IntFlag{
			Name:        "port",
			Aliases:     []string{"p"},
			Usage:       "port that the tool will use for discovery purposes",
			Value:       9045,
			EnvVars:     []string{"RAGNO_PORT"},
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
	case <-ctx.Context.Done():
		logrus.Error("Context died")
		return nil

	case <-closeC:
		logrus.Info("Shutdown detected")
		doneC <- struct{}{}
		wg.Wait()
		return nil
	}
}

func runDiscv4Service(ctx *cli.Context, wg *sync.WaitGroup, doneC chan struct{}, port int, output string) error {
	discv4, err := peerdiscovery.NewDiscv4(port)
	if err != nil {
		return err
	}

	enrC, err := discv4.Run()
	if err != nil {
		return err
	}

	csvExp, err := csvs.NewCsvExporter(output, models.ENR{}.CSVheaders())
	if err != nil {
		return err
	}

	// compose the nodeset
	enrSet := models.NewEnodeSet()

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
		for {
			select {
			case <-ctx.Context.Done():
				logrus.Error("unhandled ctx received")
				return
			case <-doneC:
				logrus.Info("Shutdown detected")
				return
			case node := <-enrC:
				logrus.WithFields(logrus.Fields{
					"enr":    node.Node.String(),
					"ID":     node.ID,
					"IP":     node.IP,
					"UDP":    node.UDP,
					"TCP":    node.TCP,
					"seq":    node.Seq,
					"pubkey": node.Pubkey,
				}).Debug("new discv4 node")
				err = enrSet.AddNode(node)
				if err != nil {
					logrus.Error(errors.Wrap(err, "unable to store new node"))
				}
			}
		}
	}()

	// persister routine
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			csvExp.Close()
			logrus.Info("discv4 persister down")
		}()
		persistT := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-closeC:
				logrus.Info("shutdown detected (persister)")
				err := csvExp.Export(enrSet.PeerRows(), enrSet.RowComposer)
				if err != nil {
					log.Error("unable to export ENR-Set", err.Error())
				}
				return
			case <-persistT.C:
				logrus.Infof("flushing peer list (%d peers) to csv", enrSet.Len())
				err := csvExp.Export(enrSet.PeerRows(), enrSet.RowComposer)
				if err != nil {
					log.Error("unable to export ENR-Set", err.Error())
				}
				persistT.Reset(30 * time.Second)
			}
		}
	}()
	return nil
}
