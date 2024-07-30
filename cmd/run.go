package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"

	"github.com/cortze/ragno/crawler"
)

var RunCommand = &cli.Command{
	Name:   "run",
	Usage:  "Run spawns an Ethereum EL crawler and starts discovering and identifying them",
	Action: RunRagno,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "log-level",
			Usage:       "Define the log level of the logs it will display on the terminal",
			EnvVars:     []string{"LOG_LEVEL"},
			DefaultText: crawler.DefaultLogLevel,
		},
		&cli.StringFlag{
			Name:        "db-endpoint",
			Usage:       "Endpoint of the database that where the results of the crawl will be stored (needs to be initialized from before)",
			EnvVars:     []string{"DB_ENDPOINT"},
			DefaultText: crawler.DefaultDBEndpoint,
		},
		&cli.StringFlag{
			Name:        "ip",
			Usage:       "IP that will be assigned to the host",
			EnvVars:     []string{"IP"},
			DefaultText: crawler.DefaultHostIP,
		},
		&cli.IntFlag{
			Name:    "port",
			Usage:   "Port that will be used by the crawler to establish TCP connections with the rest of the network",
			EnvVars: []string{"PORT"},
		},
		&cli.StringFlag{
			Name:        "metrics-ip",
			Usage:       "IP where the metrics of the crawler will be shown into",
			EnvVars:     []string{"METRICS_IP"},
			DefaultText: crawler.DefaultMetricsIP,
		},
		&cli.IntFlag{
			Name:    "metrics-port",
			Usage:   "Port that will be used to expose pprof and prometheus metrics",
			EnvVars: []string{"METRICS_PORT"},
		},
		&cli.StringFlag{
			Name:    "metrics-endpoint",
			Usage:   "Name of the endpoint where metrics will be served",
			EnvVars: []string{"METRICS_ENDPOINT"},
		},
		&cli.StringFlag{
			Name:    "dialers",
			Usage:   "Number of workers that will be used to connect to the nodes",
			Aliases: []string{"cd"},
			EnvVars: []string{"DIALERS"},
		},
		&cli.StringFlag{
			Name:    "persisters",
			Usage:   "Number of workers that will be used to save into the DB",
			Aliases: []string{"cs"},
			EnvVars: []string{"SAVER_NUM"},
		},
		&cli.StringFlag{
			Name:    "conn-timeout",
			Usage:   "Timeout in seconds for peer connection",
			Aliases: []string{"ct"},
			EnvVars: []string{"CONN_TIMEOUT"},
		},
		&cli.StringFlag{
			Name:    "snapshot-interval",
			Usage:   "Time string for how often active_peers snapshots are taken",
			Aliases: []string{"si"},
			EnvVars: []string{"SNAPSHOT_INTERVAL"},
		},
	},
}

func RunRagno(ctx *cli.Context) error {
	mainCtx, cancel := context.WithCancel(ctx.Context)
	defer cancel()

	// create a default crawler.ration
	conf := crawler.NewDefaultRun()
	err := conf.Apply(ctx)
	if err != nil {
		return errors.Wrap(err, "error applying the received configuration")
	}

	// create a new crawler from the given configuration1
	ragno, err := crawler.NewCrawler(mainCtx, *conf)
	if err != nil {
		return errors.Wrap(err, "error initializing the crawler")
	}

	// create a routine that will check whether the program needs to shut down
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGKILL, os.Interrupt)
	go func() {
		sig := <-sigChan
		log.Warnf("received signal %s - stopping ragno", sig.String())
		ragno.Close()
		close(sigChan)
	}()

	// start the crawler
	return ragno.Run()
}
