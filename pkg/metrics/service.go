package metrics

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type PrometheusMetrics struct {
	ctx context.Context

	IP              string
	port            string
	endpoint        string
	refreshInterval time.Duration

	modules []*MetricsModule

	wg     sync.WaitGroup
	closeC chan struct{}
}

func NewPrometheusMetrics(
	ctx context.Context,
	IP string,
	port int,
	endpoint string,
	refreshInterval time.Duration,
) *PrometheusMetrics {
	return &PrometheusMetrics{
		ctx:             ctx,
		IP:              IP,
		port:            fmt.Sprintf("%d", port),
		endpoint:        endpoint,
		refreshInterval: refreshInterval,
		modules:         make([]*MetricsModule, 0),
		closeC:          make(chan struct{}),
	}
}

func (pMetrics *PrometheusMetrics) AddMetricsModule(module *MetricsModule) {
	pMetrics.modules = append(pMetrics.modules, module)
}

func (pMetrics *PrometheusMetrics) Start() error {
	http.Handle("/"+pMetrics.endpoint, promhttp.Handler())

	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", pMetrics.IP, pMetrics.port), nil))
	}()

	err := pMetrics.initPrometheusMetrics()
	if err != nil {
		return errors.Wrap(err, "Unable to initialize Prometheus Metrics.")
	}

	pMetrics.wg.Add(1)
	go pMetrics.launchMetricsUpdater()

	return nil
}

func (pMetrics *PrometheusMetrics) initPrometheusMetrics() error {
	log.Debugf("Initializing %d Metric modules", len(pMetrics.modules))

	for _, module := range pMetrics.modules {
		err := module.Init()
		if err != nil {
			return err
		}
	}
	return nil
}

func (pMetrics *PrometheusMetrics) updateSubmodules() {
	log.Trace("Updating values for Prometheus Metrics")

	for _, mod := range pMetrics.modules {
		summary := make(map[string]interface{}, 0)
		moduleSummary := mod.UpdateSummary()

		for key, value := range moduleSummary {
			summary[key] = value
		}
		// compose a message with the give summary
		logFields := log.Fields(moduleSummary)
		log.WithFields(logFields).Infof("Summary for %s", mod.Name())
	}
}

func (pMetrics *PrometheusMetrics) launchMetricsUpdater() {
	defer pMetrics.wg.Done()

	ticker := time.NewTicker(pMetrics.refreshInterval)

metricsUpdateLoop:
	for {
		select {
		case <-ticker.C:
			pMetrics.updateSubmodules()

		case <-pMetrics.closeC:
			log.Debug("Detected a controled shutdown")
			break metricsUpdateLoop

		case <-pMetrics.ctx.Done():
			log.Debug("Detected that context died, shutting down")
			break metricsUpdateLoop
		}
	}

}

func (pMetrics *PrometheusMetrics) Close() {
	log.Debugf("Closing %d prometheus metrics modules", len(pMetrics.modules))
	pMetrics.closeC <- struct{}{}
	pMetrics.wg.Wait()
	log.Debug("Prometheus Metrics exporter successfully closed.")
}
