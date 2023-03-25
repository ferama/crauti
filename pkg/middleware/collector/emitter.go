package collector

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/ferama/crauti/pkg/middleware/cache"
	"github.com/ferama/crauti/pkg/middleware/proxy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type emitterMiddleware struct {
	next http.Handler

	metricPathKey string
}

func NewEmitterrMiddleware(next http.Handler, mountPointPath string) http.Handler {
	if mountPointPath != "" {
		MetricsInstance().RegisterMountPath(mountPointPath)
	}
	m := &emitterMiddleware{
		next:          next,
		metricPathKey: mountPointPath,
	}
	return m
}

func (m *emitterMiddleware) emitLogs(r *http.Request) {
	collectorContext := r.Context().Value(collectorContextKey).(collectorContext)

	totalLatency := time.Since(collectorContext.StartTime)

	uri := r.URL.Path
	if r.URL.RawQuery != "" {
		uri = fmt.Sprintf("%s?%s", uri, r.URL.RawQuery)
	}
	remoteAddr, _, _ := net.SplitHostPort(r.RemoteAddr)

	httpRequestDict := zerolog.Dict().
		Str("method", r.Method).
		Str("host", r.Host).
		Str("uri", uri).
		Int("status", collectorContext.ResponseWriter.Status()).
		Int64("requestSize", r.ContentLength).
		Int("responseSize", collectorContext.ResponseWriter.BytesWritten()).
		Str("userAgent", r.UserAgent()).
		Str("remoteIp", remoteAddr).
		Str("referer", r.Referer()).
		Float64("latency", totalLatency.Seconds()).
		Str("latency_human", totalLatency.Round(1*time.Millisecond).String()).
		Str("protocol", r.Proto)

	event := log.Info().
		Dict("httpRequest", httpRequestDict)

	cacheContext := r.Context().Value(cache.CacheContextKey)
	if cacheContext != nil {
		event.Str("cache", cacheContext.(cache.CacheContext).Status)
	}

	proxyContext := r.Context().Value(proxy.ProxyContextKey)
	if proxyContext != nil {
		pc := proxyContext.(proxy.ProxyContext)
		upstream := fmt.Sprintf("%s:%s", pc.Upstream.Hostname(), pc.Upstream.Port())
		upstreamLatency := time.Since(pc.UpstreamRequestStartTime)

		proxyUpstreamDict := zerolog.Dict().
			Str("host", upstream).
			Str("mountPath", pc.MountPath).
			Float64("latency", upstreamLatency.Seconds()).
			Str("latency_human", upstreamLatency.Round(1*time.Millisecond).String())

		event.Dict("proxyUpstream", proxyUpstreamDict)

	}

	event.Send()
}

func (m *emitterMiddleware) emitMetrics(r *http.Request) {
	collectorContext := r.Context().Value(collectorContextKey).(collectorContext)

	// status counter
	s := collectorContext.ResponseWriter.Status()
	key := MetricsInstance().GetProcessedTotalMapKey(m.metricPathKey, s)
	c, ok := MetricsInstance().Get(key)
	if ok {
		c.(prometheus.Counter).Inc()
	}

	// request latency
	totalLatency := time.Since(collectorContext.StartTime).Seconds()
	key = MetricsInstance().GetRequestLatencyMapKey(m.metricPathKey)
	c, ok = MetricsInstance().Get(key)
	if ok {
		c.(prometheus.Observer).Observe(totalLatency)
	}

	// upstream request latency
	proxyContext := r.Context().Value(proxy.ProxyContextKey)
	if proxyContext != nil {
		pc := proxyContext.(proxy.ProxyContext)
		upstreamLatency := time.Since(pc.UpstreamRequestStartTime).Seconds()

		key = MetricsInstance().GetUpstreamRequestLatencyMapKey(m.metricPathKey)
		c, ok = MetricsInstance().Get(key)
		if ok {
			c.(prometheus.Observer).Observe(upstreamLatency)
		}
	}
}

func (m *emitterMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	m.emitLogs(r)
	m.emitMetrics(r)
	m.next.ServeHTTP(w, r)
}
