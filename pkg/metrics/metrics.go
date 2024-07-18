package metrics

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type MetricsModule struct {
	name    string
	details string

	Metrics []*Metric
}

func NewMetricsModule(
	name string, details string,
) *MetricsModule {

	module := &MetricsModule{
		name:    name,
		details: details,
	}
	return (module)
}

func (metrics *MetricsModule) AddMetric(metric *Metric) {
	metrics.Metrics = append(metrics.Metrics, metric)
}

func (metrics *MetricsModule) Init() error {
	for _, metric := range metrics.Metrics {
		err := metric.Init()

		if err != nil {
			return errors.Wrap(err, "Error registering metric "+metric.Name())
		}
	}
	return (nil)
}

func (metrics *MetricsModule) UpdateSummary() map[string]interface{} {
	summary := make(map[string]interface{}, 0)

	for _, metric := range metrics.Metrics {
		metricSummary, err := metric.Update()
		if err != nil {
			log.Error("Unable to update metrics for individual metric " + metric.Name())
		}
		summary[metric.Name()] = metricSummary
	}
	return (summary)
}

func (metrics *MetricsModule) Name() string {
	return (metrics.name)
}

func (metrics *MetricsModule) Details() string {
	return (metrics.details)
}

type Metric struct {
	name       string
	initFunc   func() error
	updateFunc func() (interface{}, error)
}

func NewMetric(
	name string, initFunc func() error,
	updateFunc func() (interface{}, error),
) *Metric {
	module := &Metric{
		name:       name,
		initFunc:   initFunc,
		updateFunc: updateFunc,
	}
	return (module)
}

func (metric *Metric) Init() error {
	log.Infof("Initializing exporter %s", metric.name)
	return (metric.initFunc())
}

func (metric *Metric) Update() (interface{}, error) {
	return (metric.updateFunc())
}

func (metric *Metric) Name() string {
	return (metric.name)
}
