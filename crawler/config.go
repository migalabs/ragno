package crawler

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var ()

type CrawlerRunConf struct {
	LogLevel          string `yaml:"log-level"`
	DbEndpoint        string `yaml:"db-endpoint"`
	HostIP            string `yaml:"ip"`
	HostPort          int    `yaml:"port"`
	MetricsIP         string `yaml:"metrics-ip"`
	MetricsPort       int    `yaml:"metics-port"`
	File              string `yaml:"csv-file"`
	ConcurrentDialers int    `yaml:"concurrent-dialers"`
	ConcurrentSavers  int    `yaml:"concurrent-savers"`
}

func NewDefaultRun() *CrawlerRunConf {
	return &CrawlerRunConf{
		LogLevel:          DefaultLogLevel,
		DbEndpoint:        DefaultDBEndpoint,
		HostIP:            DefaultHostIP,
		HostPort:          DefaultHostPort,
		MetricsIP:         DefaultMetricsIP,
		MetricsPort:       DefaultMetricsPort,
		ConcurrentDialers: DefaultConcurrentDialers,
		ConcurrentSavers:  DefaultConcurrentSavers,
	}
}

// Only considered the configuration for the Execution Layer's crawler -> RunCommand
func (c *CrawlerRunConf) Apply(ctx *cli.Context) error {
	if ctx.IsSet("log-level") {
		parsedLevel, err := logrus.ParseLevel(ctx.String("log-level"))
		if err != nil {
			logrus.Warnf("invalid log level %s, using %s", ctx.String("log-level"), DefaultLogLevel)
		} else {
			c.LogLevel = parsedLevel.String()
			logrus.SetLevel(parsedLevel)
		}
	}
	if ctx.IsSet("db-endpoint") {
		c.DbEndpoint = ctx.String("db-endpoint")
	}
	if ctx.IsSet("ip") {
		c.HostIP = ctx.String("ip")
	}
	if ctx.IsSet("port") {
		c.HostPort = ctx.Int("port")
	}
	if ctx.IsSet("metrics-ip") {
		c.MetricsIP = ctx.String("metrics-ip")
	}
	if ctx.IsSet("metrics-port") {
		c.MetricsPort = ctx.Int("metrics-port")
	}
	if ctx.IsSet("file") {
		c.File = ctx.String("file")
	}
	if ctx.IsSet("concurrent-dialers") {
		c.ConcurrentDialers = ctx.Int("concurrent-dialers")
	}
	if ctx.IsSet("concurrent-savers") {
		c.ConcurrentSavers = ctx.Int("concurrent-savers")
	}
	return nil
}
