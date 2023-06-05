package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/cortze/ragno/crawler"

	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"

	"github.com/pkg/errors"
)

var RunCommand = &cli.Command{
	Name: "run", 
	Usage: "Run spawns an Ethereum EL crawler and starts discovering and identifying them",
	Action: RunRagno,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "log-level",
			Usage: "Define the log level of the logs it will display on the terminal",
			EnvVars: []string{"RAGNO_LOG_LEVEL"},
			DefaultText: crawler.DefaultLogLevel,
		},
		&cli.StringFlag{
			Name: "db-endpoint",
			Usage: "Endpoint of the database that where the results of the crawl will be stored (needs to be initialized from before)",
			EnvVars: []string{"RAGNO_DB_ENDPOINT"},
			DefaultText: crawler.DefaultDBEndpoint,
		},
		&cli.StringFlag{
			Name: "ip",
			Usage: "IP that will be assigned to the host",
			EnvVars: []string{"RAGNO_HOST_IP"},
			DefaultText: crawler.DefaultHostIP,
		},
		&cli.IntFlag{
			Name: "port",
			Usage: "Port that will be used by the crawler to stablish TCP connections with the rest of the network",
			EnvVars: []string{"RAGNO_HOST_PORT"},
		},
		&cli.StringFlag{
			Name: "metrics-ip",
			Usage: "IP where the metrics of the crawler will be shown into",
			EnvVars: []string{"RAGNO_METRICS_IP"},
			DefaultText: crawler.DefaultMetricsIP,
		},
		&cli.IntFlag{
			Name: "metrics-port",
			Usage: "Port that will be used to expose pprof and prometheus metrics",
			EnvVars: []string{"RAGNO_METRICS_PORT"},
		},
	},
}


func RunRagno(ctx *cli.Context) error {
	// create a default crawler.ration

	conf := crawler.NewDefaultRun()
	err := conf.Apply(ctx)
	if err != nil {
		return errors.Wrap(err, "error applying the received configuration")
	}

	// create a new crawler from the given configuration1
	ragno, err := crawler.NewCrawler(ctx.Context, *conf)
	if err != nil {
		return errors.Wrap(err, "error initializing the crawler")
	}	

	// wait untill the process is stoped to close it down
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)

	// run untill we receive the shutdown
	sig := <- sigs
	log.Infof("Received %s signal - Stopping Ragno with control...\n", sig.String())

	ragno.Close()


	return nil
}
