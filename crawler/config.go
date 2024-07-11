package crawler

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	// crawler host related metrics
	DefaultLogLevel             = "info"
	DefaultDBEndpoint           = "postgresql://user:password@localhost:5440/ragnodb?sslmode=disable"
	DefaultHostIP               = "0.0.0.0"
	DefaultHostPort             = 9050
	DefaultMetricsIP            = "localhost"
	DefaultMetricsPort          = 9070
	DefaultConcurrentDialers    = 150
	DefaultConcurrentPersisters = 2
	DefaultConnTimeout          = 20 * time.Second
)

type CrawlerRunConf struct {
	LogLevel    string           `yaml:"log-level"`
	DbEndpoint  string           `yaml:"db-endpoint"`
	HostIP      string           `yaml:"ip"`
	HostPort    int              `yaml:"port"`
	MetricsIP   string           `yaml:"metrics-ip"`
	MetricsPort int              `yaml:"metics-port"`
	Dialers     int              `yaml:"dialers"`
	Persisters  int              `yaml:"persisters"`
	ConnTimeout time.Duration    `yaml:"conn-timeout"`
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
		Persisters:  DefaultConcurrentPersisters,
		ConnTimeout: DefaultConnTimeout,
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
	if ctx.IsSet("conn-timeout") {
		c.ConnTimeout = time.Duration(ctx.Int("conn-timeout")) * time.Second
	}
	return nil
}
