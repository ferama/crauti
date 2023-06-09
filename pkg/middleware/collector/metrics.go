package collector

import (
	"fmt"
	"sync"

	"github.com/ferama/crauti/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	once     sync.Once
	instance *metrics
)

const (
	CrautiProcessedRequestsTotal = "crauti_processed_requests_total"
	CrautiRequestLatency         = "crauti_request_latency"
	CrautiUpstreamRequestLatency = "crauti_upstream_request_latency"
	CrautiCacheTotal             = "crauti_cache_total"
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

func (m *metrics) GetProcessedTotalMapKey(mountPath string, code int, matchHost string) string {
	var mapKey string
	if code >= 200 && code <= 299 {
		mapKey = fmt.Sprintf("%s_%s_%d_%s", CrautiProcessedRequestsTotal, mountPath, 200, matchHost)
	}
	if code >= 400 && code <= 499 {
		mapKey = fmt.Sprintf("%s_%s_%d_%s", CrautiProcessedRequestsTotal, mountPath, 400, matchHost)
	}
	if code >= 500 && code <= 599 {
		mapKey = fmt.Sprintf("%s_%s_%d_%s", CrautiProcessedRequestsTotal, mountPath, 500, matchHost)
	}
	return mapKey
}

func (m *metrics) GetRequestLatencyMapKey(mountPath string, matchHost string) string {
	mapKey := fmt.Sprintf("%s_%s_%s", CrautiRequestLatency, mountPath, matchHost)
	return mapKey
}

func (m *metrics) GetUpstreamRequestLatencyMapKey(mountPath string, matchHost string) string {
	mapKey := fmt.Sprintf("%s_%s_%s", CrautiUpstreamRequestLatency, mountPath, matchHost)
	return mapKey
}

func (m *metrics) GetCacheTotalMapKey(mountPath string, cacheStatus string, matchHost string) string {
	mapKey := fmt.Sprintf("%s_%s_%s_%s", CrautiCacheTotal, mountPath, cacheStatus, matchHost)
	return mapKey
}

// Register per mountPath prometheus metrics
func (m *metrics) RegisterMountPath(mountPath string, upstream string, matchHost string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Histogram
	// https://prometheus.io/docs/concepts/metric_types/

	////////////
	//
	// Processed within response code
	//
	code := 200
	mapKey := m.GetProcessedTotalMapKey(mountPath, code, matchHost)
	m.collectors[mapKey] = promauto.NewCounter(prometheus.CounterOpts{
		Name: CrautiProcessedRequestsTotal,
		Help: "Total processed requests",
		ConstLabels: prometheus.Labels{
			"code": fmt.Sprint(code), "mountPath": mountPath, "upstream": upstream, "host": matchHost},
	})

	code = 400
	mapKey = m.GetProcessedTotalMapKey(mountPath, code, matchHost)
	m.collectors[mapKey] = promauto.NewCounter(prometheus.CounterOpts{
		Name: CrautiProcessedRequestsTotal,
		Help: "Total processed requests",
		ConstLabels: prometheus.Labels{
			"code": fmt.Sprint(code), "mountPath": mountPath, "upstream": upstream, "host": matchHost},
	})

	code = 500
	mapKey = m.GetProcessedTotalMapKey(mountPath, code, matchHost)
	m.collectors[mapKey] = promauto.NewCounter(prometheus.CounterOpts{
		Name: CrautiProcessedRequestsTotal,
		Help: "Total processed requests",
		ConstLabels: prometheus.Labels{
			"code": fmt.Sprint(code), "mountPath": mountPath, "upstream": upstream, "host": matchHost},
	})

	////////////
	//
	// latency
	//
	// Query example:
	//  rate(crauti_request_latency_bucket{mountPath="/mount1"}[1m])
	mapKey = m.GetRequestLatencyMapKey(mountPath, matchHost)
	m.collectors[mapKey] = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:        CrautiRequestLatency,
		Help:        "Request latency",
		ConstLabels: prometheus.Labels{"mountPath": mountPath, "upstream": upstream, "host": matchHost},
		Buckets:     []float64{0.3, 0.5, 3},
	})

	mapKey = m.GetUpstreamRequestLatencyMapKey(mountPath, matchHost)
	m.collectors[mapKey] = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:        CrautiUpstreamRequestLatency,
		Help:        "Proxy upstream request latency",
		ConstLabels: prometheus.Labels{"mountPath": mountPath, "upstream": upstream, "host": matchHost},
		Buckets:     []float64{0.3, 0.5, 3},
	})

	////////////
	//
	// cache
	// Query example:
	//   1 - (sum(crauti_cache_total{status!="HIT"}) / sum(crauti_cache_total{status="HIT"}))
	mapKey = m.GetCacheTotalMapKey(mountPath, utils.CacheStatusBypass, matchHost)
	m.collectors[mapKey] = promauto.NewCounter(prometheus.CounterOpts{
		Name: CrautiCacheTotal,
		Help: "Total cache",
		ConstLabels: prometheus.Labels{
			"status": utils.CacheStatusBypass, "mountPath": mountPath, "upstream": upstream, "host": matchHost},
	})
	mapKey = m.GetCacheTotalMapKey(mountPath, utils.CacheStatusHit, matchHost)
	m.collectors[mapKey] = promauto.NewCounter(prometheus.CounterOpts{
		Name: CrautiCacheTotal,
		Help: "Total cache",
		ConstLabels: prometheus.Labels{
			"status": utils.CacheStatusHit, "mountPath": mountPath, "upstream": upstream, "host": matchHost},
	})
	mapKey = m.GetCacheTotalMapKey(mountPath, utils.CacheStatusIgnored, matchHost)
	m.collectors[mapKey] = promauto.NewCounter(prometheus.CounterOpts{
		Name: CrautiCacheTotal,
		Help: "Total cache",
		ConstLabels: prometheus.Labels{
			"status": utils.CacheStatusIgnored, "mountPath": mountPath, "upstream": upstream, "host": matchHost},
	})
	mapKey = m.GetCacheTotalMapKey(mountPath, utils.CacheStatusMiss, matchHost)
	m.collectors[mapKey] = promauto.NewCounter(prometheus.CounterOpts{
		Name: CrautiCacheTotal,
		Help: "Total cache",
		ConstLabels: prometheus.Labels{
			"status": utils.CacheStatusMiss, "mountPath": mountPath, "upstream": upstream, "host": matchHost},
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
