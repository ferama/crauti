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

type logEmitterMiddleware struct {
	next http.Handler

	metricPathKey string
}

func NewLogEmitterrMiddleware(next http.Handler, mountPointPath string) http.Handler {
	if mountPointPath != "" {
		MetricsInstance().RegisterMountPath(mountPointPath)
	}
	m := &logEmitterMiddleware{
		next:          next,
		metricPathKey: mountPointPath,
	}
	return m
}

func (m *logEmitterMiddleware) emitLogs(r *http.Request) {
	logContext := r.Context().Value(collectorContextKey).(collectorContext)

	totalLatency := time.Since(logContext.StartTime).Round(1 * time.Millisecond).Seconds()

	uri := r.URL.Path
	if r.URL.RawQuery != "" {
		uri = fmt.Sprintf("%s?%s", uri, r.URL.RawQuery)
	}
	remoteAddr, _, _ := net.SplitHostPort(r.RemoteAddr)

	httpRequestDict := zerolog.Dict().
		Str("method", r.Method).
		Str("host", r.Host).
		Str("uri", uri).
		Int("status", logContext.ResponseWriter.Status()).
		Int64("requestSize", r.ContentLength).
		Int("responseSize", logContext.ResponseWriter.BytesWritten()).
		Str("userAgent", r.UserAgent()).
		Str("remoteIp", remoteAddr).
		Str("referer", r.Referer()).
		Float64("latency", totalLatency).
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
		upstreamLatency := time.Since(pc.UpstreamRequestStartTime).Round(1 * time.Millisecond).Seconds()

		proxyUpstreamDict := zerolog.Dict().
			Str("host", upstream).
			Str("mountPath", pc.MountPath).
			Float64("latency", upstreamLatency)

		event.Dict("proxyUpstream", proxyUpstreamDict)

	}

	event.Send()
}

func (m *logEmitterMiddleware) emitMetrics(r *http.Request) {
	logContext := r.Context().Value(collectorContextKey).(collectorContext)
	s := logContext.ResponseWriter.Status()
	if s >= 200 && s <= 299 {
		key := MetricsInstance().GetProcessedTotalMapKey(m.metricPathKey, "200")
		c, ok := MetricsInstance().Get(key)
		if ok {
			c.(prometheus.Counter).Inc()
		}
	}
	if s >= 400 && s <= 499 {
		key := MetricsInstance().GetProcessedTotalMapKey(m.metricPathKey, "400")
		c, ok := MetricsInstance().Get(key)
		if ok {
			c.(prometheus.Counter).Inc()
		}
	}
	if s >= 500 && s <= 599 {
		key := MetricsInstance().GetProcessedTotalMapKey(m.metricPathKey, "500")
		c, ok := MetricsInstance().Get(key)
		if ok {
			c.(prometheus.Counter).Inc()
		}
	}
}

func (m *logEmitterMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	m.emitLogs(r)
	m.emitMetrics(r)
	m.next.ServeHTTP(w, r)
}
