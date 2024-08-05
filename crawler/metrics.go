package crawler

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/cortze/ragno/pkg/metrics"
)

var (
	moduleName    = "crawler"
	moduleDetails = "General Ragno metrics"

	// List of metrics that we are going to export
	ClientDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: moduleName,
		Name:      "client_distribution",
		Help:      "Number of peers using clients seen",
	},
		[]string{"client"},
	)
	VersionDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: moduleName,
		Name:      "observed_client_version_distribution",
		Help:      "Number of peers from each of the clients versions",
	},
		[]string{"client_version"},
	)
	GeoDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: moduleName,
		Name:      "geographical_distribution",
		Help:      "Number of peers from each country",
	},
		[]string{"country"},
	)
	NodeDistribution = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: moduleName,
		Name:      "node_distribution",
		Help:      "Number of peers from each of the crawled countries",
	})
	DeprecatedCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: moduleName,
		Name:      "deprecated_nodes",
		Help:      "Total number of deprecated peers",
	})
	OsDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: moduleName,
		Name:      "os_distribution",
		Help:      "OS distribution of connected peers",
	},
		[]string{"os"},
	)
	ArchDistribution = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: moduleName,
		Name:      "arch_distribution",
		Help:      "Architecture distribution of the active peers in the network",
	},
		[]string{"arch"},
	)
	HostedPeers = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: moduleName,
		Name:      "hosted_peers_distribution",
		Help:      "Distribution of nodes that are hosted on non-residential networks",
	},
		[]string{"ip_host"},
	)
	// RttDist = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	// 	Namespace: moduleName,
	// 	Name:      "observed_rtt_distribution",
	// 	Help:      "Distribution of RTT between the crawler and nodes in the network",
	// },
	// 	[]string{"secs"},
	// )
	IPDist = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: moduleName,
		Name:      "observed_ip_distribution",
		Help:      "Distribution of IPs hosting nodes in the network",
	},
		[]string{"numbernodes"},
	)
)

func (crawler *Crawler) GetMetrics() *metrics.MetricsModule {
	metricsModule := metrics.NewMetricsModule(
		moduleName,
		moduleDetails,
	)

	metricsModule.AddMetric(crawler.GetClientDistributionMetrics())
	metricsModule.AddMetric(crawler.versionDistributionMetrics())
	metricsModule.AddMetric(crawler.geoDistributionMetrics())
	metricsModule.AddMetric(crawler.nodeDistributionMetrics())
	metricsModule.AddMetric(crawler.deprecatedNodeMetrics())
	metricsModule.AddMetric(crawler.getPeersOs())
	metricsModule.AddMetric(crawler.getPeersArch())
	metricsModule.AddMetric(crawler.getHostedPeers())
	// metricsModule.AddMetric(crawler.getRTTDist())
	metricsModule.AddMetric(crawler.getIPDist())
	return (metricsModule)
}

func (crawler *Crawler) GetClientDistributionMetrics() *metrics.Metric {
	initFn := func() error {
		prometheus.MustRegister(ClientDistribution)
		return nil
	}
	updateFn := func() (interface{}, error) {
		summary, err := crawler.db.GetClientDistribution()
		if err != nil {
			return nil, err
		}
		for cliName, cnt := range summary {
			ClientDistribution.WithLabelValues(cliName).Set(float64(cnt.(int)))
		}
		return summary, nil
	}
	cliDist := metrics.NewMetric(
		"client_distribution",
		initFn,
		updateFn,
	)
	return (cliDist)
}

func (c *Crawler) versionDistributionMetrics() *metrics.Metric {
	initFn := func() error {
		prometheus.MustRegister(VersionDistribution)
		return nil
	}
	updateFn := func() (interface{}, error) {
		summary, err := c.db.GetVersionDistribution()
		if err != nil {
			return nil, err
		}
		for cliVer, cnt := range summary {
			VersionDistribution.WithLabelValues(cliVer).Set(float64(cnt.(int)))
		}
		return summary, nil
	}
	versDist := metrics.NewMetric(
		"client_version_distribution",
		initFn,
		updateFn,
	)
	return versDist
}

func (c *Crawler) geoDistributionMetrics() *metrics.Metric {
	initFn := func() error {
		prometheus.MustRegister(GeoDistribution)
		return nil
	}
	updateFn := func() (interface{}, error) {
		summary, err := c.db.GetGeoDistribution()
		if err != nil {
			fmt.Println(errors.Wrap(err, "unable to get GeoDist"))
			return nil, err
		}
		for country, cnt := range summary {
			GeoDistribution.WithLabelValues(country).Set(float64(cnt.(int)))
		}
		return summary, nil
	}
	versDist := metrics.NewMetric(
		"geographical_distribution",
		initFn,
		updateFn,
	)
	return versDist
}

func (c *Crawler) nodeDistributionMetrics() *metrics.Metric {
	initFn := func() error {
		prometheus.MustRegister(NodeDistribution)
		return nil
	}
	updateFn := func() (interface{}, error) {
		peerLs, err := c.db.GetNonDeprecatedNodes(c.peering.host.localChainStatus.NetworkID)
		if err != nil {
			return nil, err
		}

		NodeDistribution.Set(float64(len(peerLs)))

		return len(peerLs), nil
	}
	nodeDist := metrics.NewMetric(
		"geographical_distribution",
		initFn,
		updateFn,
	)
	return nodeDist
}

func (c *Crawler) deprecatedNodeMetrics() *metrics.Metric {
	initFn := func() error {
		prometheus.MustRegister(DeprecatedCount)
		return nil
	}
	updateFn := func() (interface{}, error) {
		nodeCnt, err := c.db.GetDeprecatedNodes()
		if err != nil {
			return nil, err
		}
		DeprecatedCount.Set(float64(nodeCnt))
		return nodeCnt, nil
	}
	depNodes := metrics.NewMetric(
		"deprecated_nodes",
		initFn,
		updateFn,
	)
	return depNodes
}

func (c *Crawler) getPeersOs() *metrics.Metric {
	initFn := func() error {
		prometheus.MustRegister(OsDistribution)
		return nil
	}
	updateFn := func() (interface{}, error) {
		osDist, err := c.db.GetOsDistribution()
		if err != nil {
			return nil, err
		}
		for key, val := range osDist {
			OsDistribution.WithLabelValues(key).Set(float64(val.(int)))
		}
		return osDist, nil
	}
	osMetr := metrics.NewMetric(
		"os_distribution",
		initFn,
		updateFn,
	)
	return osMetr
}

func (c *Crawler) getPeersArch() *metrics.Metric {
	initFn := func() error {
		prometheus.MustRegister(ArchDistribution)
		return nil
	}
	updateFn := func() (interface{}, error) {
		archDist, err := c.db.GetArchDistribution()
		if err != nil {
			return nil, err
		}
		for key, val := range archDist {
			ArchDistribution.WithLabelValues(key).Set(float64(val.(int)))
		}
		return archDist, nil
	}
	archMetr := metrics.NewMetric(
		"arch_distribution",
		initFn,
		updateFn,
	)
	return archMetr
}

func (c *Crawler) getHostedPeers() *metrics.Metric {
	initFn := func() error {
		prometheus.MustRegister(HostedPeers)
		return nil
	}
	updateFn := func() (interface{}, error) {
		ipSummary, err := c.db.GetHostingDistribution()
		if err != nil {
			return nil, err
		}
		for key, val := range ipSummary {
			HostedPeers.WithLabelValues(key).Set(float64(val.(int)))
		}
		return ipSummary, nil
	}
	ipHosting := metrics.NewMetric(
		"hosted_peer_distribution",
		initFn,
		updateFn,
	)
	return ipHosting
}

// func (c *Crawler) getRTTDist() *metrics.Metric {
// 	initFn := func() error {
// 		prometheus.MustRegister(RttDist)
// 		return nil
// 	}
// 	updateFn := func() (interface{}, error) {
// 		summary, err := c.db.GetRTTDistribution()
// 		if err != nil {
// 			return nil, err
// 		}
// 		for key, val := range summary {
// 			RttDist.WithLabelValues(key).Set(float64(val.(int)))
// 		}
// 		return summary, nil
// 	}
// 	indvMetric := metrics.NewMetric(
// 		"rtt_distribution",
// 		initFn,
// 		updateFn,
// 	)
// 	return indvMetric
// }

func (c *Crawler) getIPDist() *metrics.Metric {
	initFn := func() error {
		prometheus.MustRegister(IPDist)
		return nil
	}
	updateFn := func() (interface{}, error) {
		summary, err := c.db.GetIPDistribution()
		if err != nil {
			return nil, err
		}
		for key, val := range summary {
			IPDist.WithLabelValues(key).Set(float64(val.(int)))
		}
		return summary, nil
	}
	indvMetric := metrics.NewMetric(
		"ip_distribution",
		initFn,
		updateFn,
	)
	return indvMetric
}
