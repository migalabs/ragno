package crawler

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	// crawler host related metrics
	DefaultLogLevel          = "info"
	DefaultDBEndpoint        = "postgresql://user:password@localhost:5432/ragnodb"
	DefaultDiscPort          = 9045
	DefaultHostIP            = "0.0.0.0"
	DefaultHostPort          = 9050
	DefaultMetricsIP         = "localhost"
	DefaultMetricsPort       = 9070
	DefaultConcurrentDialers = 150
	DefaultConcurrentSavers  = 2
	DefaultRetryAmount       = 3
	DefaultRetryDelay        = 8

	// Not using yaml files so far
	DefaultConfigFile = "config/example.yaml"
)

type CrawlerRunConf struct {
	LogLevel    string `yaml:"log-level"`
	DbEndpoint  string `yaml:"db-endpoint"`
	HostIP      string `yaml:"ip"`
	HostPort    int    `yaml:"port"`
	MetricsIP   string `yaml:"metrics-ip"`
	MetricsPort int    `yaml:"metics-port"`
	Dialers     int    `yaml:"dialers"`
	Persisters  int    `yaml:"persisters"`
	Retries     int    `yaml:"retries"`
}

func NewDefaultRun() *CrawlerRunConf {
	return &CrawlerRunConf{
		LogLevel:    DefaultLogLevel,
		DbEndpoint:  DefaultDBEndpoint,
		HostIP:      DefaultHostIP,
		HostPort:    DefaultHostPort,
		MetricsIP:   DefaultMetricsIP,
		MetricsPort: DefaultMetricsPort,
		Dialers:     DefaultConcurrentDialers,
		Persisters:  DefaultConcurrentSavers,
		Retries:     DefaultRetryAmount,
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
	if ctx.IsSet("dialers") {
		c.Dialers = ctx.Int("dialers")
	}
	if ctx.IsSet("persisters") {
		c.Persisters = ctx.Int("persisters")
	}
	if ctx.IsSet("retries") {
		c.Retries = ctx.Int("retries")
	}

	return nil
}
