package collector

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
)

var (
	once     sync.Once
	instance *metrics
)

const (
	CrautiProcessedRequestsTotal = "crauti_processed_requests_total"
	CrautiRequestLatency         = "crauti_request_latency"
	CrautiUpstreamRequestLatency = "crauti_upstream_request_latency"
)

func MetricsInstance() *metrics {
	once.Do(func() {
		instance = newMetrics()
	})

	return instance
}

type metrics struct {
	collectors map[string]prometheus.Collector

	mu sync.Mutex
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
		mapKey = fmt.Sprintf("%s_%s_%d", CrautiProcessedRequestsTotal, mountPath, 200)
	}
	if code >= 400 && code <= 499 {
		mapKey = fmt.Sprintf("%s_%s_%d", CrautiProcessedRequestsTotal, mountPath, 400)
	}
	if code >= 500 && code <= 599 {
		mapKey = fmt.Sprintf("%s_%s_%d", CrautiProcessedRequestsTotal, mountPath, 500)
	}
	return mapKey
}

func (m *metrics) GetRequestLatencyMapKey(mountPath string) string {
	mapKey := fmt.Sprintf("%s_%s", CrautiRequestLatency, mountPath)
	return mapKey
}

func (m *metrics) GetUpstreamRequestLatencyMapKey(mountPath string) string {
	mapKey := fmt.Sprintf("%s_%s", CrautiUpstreamRequestLatency, mountPath)
	return mapKey
}

// Register per mountPath prometheus metrics
func (m *metrics) RegisterMountPath(mountPath string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Debug().Str("mountPath", mountPath).Msg("registering mount path")
	// https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Histogram
	// https://prometheus.io/docs/concepts/metric_types/

	////////////
	//
	// Processed within response code
	//
	code := 200
	mapKey := m.GetProcessedTotalMapKey(mountPath, code)
	m.collectors[mapKey] = promauto.NewCounter(prometheus.CounterOpts{
		Name:        CrautiProcessedRequestsTotal,
		Help:        "Total processed requests",
		ConstLabels: prometheus.Labels{"code": fmt.Sprint(code), "mountPath": mountPath},
	})

	code = 400
	mapKey = m.GetProcessedTotalMapKey(mountPath, code)
	m.collectors[mapKey] = promauto.NewCounter(prometheus.CounterOpts{
		Name:        CrautiProcessedRequestsTotal,
		Help:        "Total processed requests",
		ConstLabels: prometheus.Labels{"code": fmt.Sprint(code), "mountPath": mountPath},
	})

	code = 500
	mapKey = m.GetProcessedTotalMapKey(mountPath, code)
	m.collectors[mapKey] = promauto.NewCounter(prometheus.CounterOpts{
		Name:        CrautiProcessedRequestsTotal,
		Help:        "Total processed requests",
		ConstLabels: prometheus.Labels{"code": fmt.Sprint(code), "mountPath": mountPath},
	})

	////////////
	//
	// latency
	//
	// Query example:
	//  rate(crauti_request_latency_bucket{mountPath="/mount1"}[1m])
	mapKey = m.GetRequestLatencyMapKey(mountPath)
	m.collectors[mapKey] = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:        CrautiRequestLatency,
		Help:        "Request latency",
		ConstLabels: prometheus.Labels{"mountPath": mountPath},
		Buckets:     []float64{0.3, 0.5, 3},
	})

	mapKey = m.GetUpstreamRequestLatencyMapKey(mountPath)
	m.collectors[mapKey] = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:        CrautiUpstreamRequestLatency,
		Help:        "Proxy upstream request latency",
		ConstLabels: prometheus.Labels{"mountPath": mountPath},
		Buckets:     []float64{0.3, 0.5, 3},
	})
}

func (m *metrics) Get(key string) (prometheus.Collector, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	val, ok := m.collectors[key]
	return val, ok
}

func (m *metrics) UnregisterAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for k, v := range m.collectors {
		prometheus.DefaultRegisterer.Unregister(v)
		delete(m.collectors, k)
	}
}
