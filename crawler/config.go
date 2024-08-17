package crawler

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	// crawler host related metrics
	DefaultLogLevel             = "info"
	DefaultDBEndpoint           = "postgresql://user:password@localhost:5440"
	DefaultHostIP               = "0.0.0.0"
	DefaultHostPort             = 9050
	DefaultMetricsIP            = "localhost"
	DefaultMetricsPort          = 9070
	DefaultMetricsEndpoint      = "metrics"
	DefaultConcurrentDialers    = 150
	DefaultConcurrentPersisters = 2
	DefaultConnTimeout          = 30 * time.Second
	DefaultSnapshotInterval     = 30 * time.Minute
	DefaultIPAPIUrl             = "http://ip-api.com/json/{__ip__}?fields=status,continent,continentCode,country,countryCode,region,regionName,city,zip,lat,lon,isp,org,as,asname,mobile,proxy,hosting,query"
	DefaultDeprecationTime      = 48 * time.Hour
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
	DeprecationTime  time.Duration `yaml:"deprecation-time"`
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
		DeprecationTime:  DefaultDeprecationTime,
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
	config := map[string]func(flag string){
		"log-level": func(flag string) {
			parsedLevel, err := logrus.ParseLevel(ctx.String("log-level"))
			if err != nil {
				logrus.Warnf("invalid log level %s, using %s", ctx.String("log-level"), DefaultLogLevel)
			} else {
				c.LogLevel = parsedLevel.String()
				logrus.SetLevel(parsedLevel)
			}
		},
		"db-endpoint":       func(flag string) { c.DbEndpoint = ctx.String(flag) },
		"ip":                func(flag string) { c.HostIP = ctx.String(flag) },
		"port":              func(flag string) { c.HostPort = ctx.Int(flag) },
		"metrics-ip":        func(flag string) { c.MetricsIP = ctx.String(flag) },
		"metrics-port":      func(flag string) { c.MetricsPort = ctx.Int(flag) },
		"metrics-endpoint":  func(flag string) { c.MetricsEndpoint = ctx.String(flag) },
		"dialers":           func(flag string) { c.Dialers = ctx.Int(flag) },
		"persisters":        func(flag string) { c.Persisters = ctx.Int(flag) },
		"conn-timeout":      func(flag string) { c.ConnTimeout = c.parseDurationVar(flag, DefaultConnTimeout, ctx) },
		"snapshot-interval": func(flag string) { c.SnapshotInterval = c.parseDurationVar(flag, DefaultSnapshotInterval, ctx) },
		"ip-api-url":        func(flag string) { c.IPAPIUrl = ctx.String(flag) },
		"deprecation-time":  func(flag string) { c.DeprecationTime = c.parseDurationVar(flag, DefaultDeprecationTime, ctx) },
	}

	for flag, applier := range config {
		if ctx.IsSet(flag) {
			applier(flag)
		}
	}
	return nil
}
