package crawler

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	// crawler host related metrics
	DefaultLogLevel             = "info"
	DefaultDBEndpoint           = "postgresql://user:password@localhost:5432/ragnodb"
	DefaultHostIP               = "0.0.0.0"
	DefaultHostPort             = 9050
	DefaultMetricsIP            = "localhost"
	DefaultMetricsPort          = 9070
	DefaultMetricsEndpoint      = "metrics"
	DefaultConcurrentDialers    = 150
	DefaultConcurrentPersisters = 2
	DefaultConnTimeout          = 30 * time.Second
	DefaultSnapshotInterval     = 12 * time.Hour
	DefaultIPAPIUrl             = "http://ip-api.com/json/{__ip__}?fields=status,continent,continentCode,country,countryCode,region,regionName,city,zip,lat,lon,isp,org,as,asname,mobile,proxy,hosting,query"
)

type CrawlerRunConf struct {
	LogLevel         string        `yaml:"log-level"`
	DbEndpoint       string        `yaml:"db-endpoint"`
	HostIP           string        `yaml:"ip"`
	HostPort         int           `yaml:"port"`
	MetricsIP        string        `yaml:"metrics-ip"`
	MetricsPort      int           `yaml:"metrics-port"`
	MetricsEndpoint  string        `yaml:"metrics-endpoint"`
	Dialers          int           `yaml:"dialers"`
	Persisters       int           `yaml:"persisters"`
	ConnTimeout      time.Duration `yaml:"conn-timeout"`
	SnapshotInterval time.Duration `yaml:"snapshot-interval"`
	IPAPIUrl         string        `yaml:"snapshot-interval"`
}

func NewDefaultRun() *CrawlerRunConf {
	return &CrawlerRunConf{
		LogLevel:         DefaultLogLevel,
		DbEndpoint:       DefaultDBEndpoint,
		HostIP:           DefaultHostIP,
		HostPort:         DefaultHostPort,
		MetricsIP:        DefaultMetricsIP,
		MetricsPort:      DefaultMetricsPort,
		MetricsEndpoint:  DefaultMetricsEndpoint,
		Dialers:          DefaultConcurrentDialers,
		Persisters:       DefaultConcurrentPersisters,
		ConnTimeout:      DefaultConnTimeout,
		SnapshotInterval: DefaultSnapshotInterval,
		IPAPIUrl:         DefaultIPAPIUrl,
	}
}

func (c *CrawlerRunConf) parseDurationVar(
	timeVar string, defaultTime time.Duration, ctx *cli.Context,
) time.Duration {
	parsedTime, parseErr := time.ParseDuration(ctx.String(timeVar))

	if parseErr != nil {
		logrus.Warnf("Interval %s is not a valid time string (%s). Using %s instead.",
			ctx.String(timeVar), parseErr, defaultTime)
		parsedTime = defaultTime
	}
	return parsedTime
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
	if ctx.IsSet("metrics-endpoint") {
		c.MetricsEndpoint = ctx.String("metrics-endpoint")
	}
	if ctx.IsSet("dialers") {
		c.Dialers = ctx.Int("dialers")
	}
	if ctx.IsSet("persisters") {
		c.Persisters = ctx.Int("persisters")
	}
	if ctx.IsSet("conn-timeout") {
		c.ConnTimeout = c.parseDurationVar("conn-timeout", DefaultConnTimeout, ctx)
	}
	if ctx.IsSet("snapshot-interval") {
		c.SnapshotInterval = c.parseDurationVar("snapshot-interval", DefaultSnapshotInterval, ctx)
	}
	if ctx.IsSet("ip-api-url") {
		c.IPAPIUrl = ctx.String("ip-api-url")
	}
	return nil
}
