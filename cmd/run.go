package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/cortze/ragno/crawler"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

var RunCommand = &cli.Command{
	Name:   "run",
	Usage:  "Run spawns an Ethereum EL crawler and starts discovering and identifying them",
	Action: RunRagno,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "log-level",
			Usage:       "Define the log level of the logs it will display on the terminal",
			EnvVars:     []string{"RAGNO_LOG_LEVEL"},
			DefaultText: crawler.DefaultLogLevel,
		},
		&cli.StringFlag{
			Name:        "db-endpoint",
			Usage:       "Endpoint of the database that where the results of the crawl will be stored (needs to be initialized from before)",
			EnvVars:     []string{"RAGNO_DB_ENDPOINT"},
			DefaultText: crawler.DefaultDBEndpoint,
		},
		&cli.StringFlag{
			Name:        "ip",
			Usage:       "IP that will be assigned to the host",
			EnvVars:     []string{"RAGNO_IP"},
			DefaultText: crawler.DefaultHostIP,
		},
		&cli.IntFlag{
			Name:    "port",
			Usage:   "Port that will be used by the crawler to establish TCP connections with the rest of the network",
			EnvVars: []string{"RAGNO_PORT"},
		},
		&cli.StringFlag{
			Name:        "metrics-ip",
			Usage:       "IP where the metrics of the crawler will be shown into",
			EnvVars:     []string{"RAGNO_METRICS_IP"},
			DefaultText: crawler.DefaultMetricsIP,
		},
		&cli.IntFlag{
			Name:    "metrics-port",
			Usage:   "Port that will be used to expose pprof and prometheus metrics",
			EnvVars: []string{"RAGNO_METRICS_PORT"},
		},
		&cli.StringFlag{
			Name:    "dialers",
			Usage:   "Number of workers that will be used to connect to the nodes",
			Aliases: []string{"cd"},
			EnvVars: []string{"RAGNO_DIALERS"},
		},
		&cli.StringFlag{
			Name:    "persisters",
			Usage:   "Number of workers that will be used to save into the DB",
			Aliases: []string{"cs"},
			EnvVars: []string{"RAGNO_SAVER_NUM"},
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
