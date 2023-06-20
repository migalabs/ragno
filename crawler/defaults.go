package crawler

import ()

var (
	// crawler host related metrics
	DefaultLogLevel    = "info"
	DefaultDBEndpoint  = "postgresql://user:password@localhost:5432/ragno"
	DefaultHostIP      = "0.0.0.0"
	DefaultHostPort    = 9050
	DefaultMetricsIP   = "localhost"
	DefaultMetricsPort = 9070
	DefaultWorkerNum   = 150
	DefaultSaverNum    = 150

	// Not using yaml files so far
	DefaultConfigFile = "config/example.yaml"
)
