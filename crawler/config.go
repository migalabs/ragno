package crawler

import (
	// "os"

	// "github.com/go-yaml/yaml"
	// "github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var ()

type CrawlerRunConf struct {
	LogLevel    string `yaml:"log-level"`
	DbEndpoint  string `yaml:"db-endpoint"`
	HostIP      string `yaml:"ip"`
	HostPort    int    `yaml:"port"`
	MetricsIP   string `yaml:"metrics-ip"`
	MetricsPort int    `yaml:"metics-port"`
	File        string `yaml:"csv-file"`
	Enr         string `yaml:"enr"`
	WorkerNum   int    `yaml:"worker-num"`
}

func NewDefaultRun() *CrawlerRunConf {
	return &CrawlerRunConf{
		LogLevel:    DefaultLogLevel,
		DbEndpoint:  DefaultDBEndpoint,
		HostIP:      DefaultHostIP,
		HostPort:    DefaultHostPort,
		MetricsIP:   DefaultMetricsIP,
		MetricsPort: DefaultMetricsPort,
		WorkerNum:   DefaultWorkerNum,
	}
}

// Only considered the configuration for the Execution Layer's crawler -> RunCommand
func (c *CrawlerRunConf) Apply(ctx *cli.Context) error {
	if ctx.IsSet("log-level") {
		c.LogLevel = ctx.String("log-level")
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
	if ctx.IsSet("enr") {
		c.Enr = ctx.String("enr")
	}
	if ctx.IsSet("worker-num") {
		c.WorkerNum = ctx.Int("worker-num")
	}
	return nil
}
