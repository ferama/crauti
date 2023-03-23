package collector

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	once     sync.Once
	instance *metrics
)

const (
	CrautiProcessedTotal = "crauti_processed_total"
)

func MetricsInstance() *metrics {
	once.Do(func() {
		instance = newMetrics()
	})

	return instance
}

type metrics struct {
	collectors map[string]prometheus.Collector
}

func newMetrics() *metrics {
	m := &metrics{
		collectors: make(map[string]prometheus.Collector),
	}
	return m
}

func (m *metrics) GetProcessedTotalMapKey(mountPath string, code int) string {
	var mapKey string
	if code >= 200 && code <= 299 {
		mapKey = fmt.Sprintf("%s_%s_%d", CrautiProcessedTotal, mountPath, 200)
	}
	if code >= 400 && code <= 499 {
		mapKey = fmt.Sprintf("%s_%s_%d", CrautiProcessedTotal, mountPath, 400)
	}
	if code >= 500 && code <= 599 {
		mapKey = fmt.Sprintf("%s_%s_%d", CrautiProcessedTotal, mountPath, 500)
	}
	return mapKey
}

// Register per mountPath prometheus metrics
func (m *metrics) RegisterMountPath(mountPath string) {
	// https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Histogram

	code := 200
	mapKey := m.GetProcessedTotalMapKey(mountPath, code)
	m.collectors[mapKey] = promauto.NewCounter(prometheus.CounterOpts{
		Name:        CrautiProcessedTotal,
		Help:        "Total processed requests",
		ConstLabels: prometheus.Labels{"code": fmt.Sprint(code), "mountPath": mountPath},
	})

	code = 400
	mapKey = m.GetProcessedTotalMapKey(mountPath, code)
	m.collectors[mapKey] = promauto.NewCounter(prometheus.CounterOpts{
		Name:        CrautiProcessedTotal,
		Help:        "Total processed requests",
		ConstLabels: prometheus.Labels{"code": fmt.Sprint(code), "mountPath": mountPath},
	})

	code = 500
	mapKey = m.GetProcessedTotalMapKey(mountPath, code)
	m.collectors[mapKey] = promauto.NewCounter(prometheus.CounterOpts{
		Name:        CrautiProcessedTotal,
		Help:        "Total processed requests",
		ConstLabels: prometheus.Labels{"code": fmt.Sprint(code), "mountPath": mountPath},
	})

}

func (m *metrics) Get(key string) (prometheus.Collector, bool) {
	val, ok := m.collectors[key]
	return val, ok
}

func (m *metrics) UnregisterAll() {
	for k, v := range m.collectors {
		prometheus.DefaultRegisterer.Unregister(v)
		delete(m.collectors, k)
	}
}
