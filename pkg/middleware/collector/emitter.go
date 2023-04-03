package collector

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/ferama/crauti/pkg/chaincontext"
	"github.com/ferama/crauti/pkg/middleware"
	"github.com/ferama/crauti/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type EmitterMiddleware struct {
	middleware.Middleware

	next http.Handler
}

func (m *EmitterMiddleware) Init(next http.Handler) middleware.Middleware {
	m.next = next
	return m
}

func (m *EmitterMiddleware) emitLogs(r *http.Request) {
	chainContext := chaincontext.GetChainContext(r)

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

	if chainContext.Conf.Middlewares.Cache.IsEnabled() {
		cacheContext := chainContext.Cache
		event.Str("cache", cacheContext.Status)
	}

	proxyContext := chainContext.Proxy
	upstreamLatency := time.Since(proxyContext.UpstreamRequestStartTime)

	proxyUpstreamDict := zerolog.Dict().
		Str("url", chainContext.Conf.Upstream).
		Str("mountPath", chainContext.Conf.Path).
		Float64("latency", upstreamLatency.Seconds()).
		Str("latency_human", upstreamLatency.Round(1*time.Millisecond).String())

	event.Dict("proxyUpstream", proxyUpstreamDict)

	event.Send()
}

func (m *EmitterMiddleware) emitMetrics(r *http.Request) {
	chainContext := chaincontext.GetChainContext(r)
	metricPathKey := chainContext.Conf.Path

	requestHost, err := utils.GetRequestHost(r)
	if err != nil {
		log.Fatal().Err(err)
	}

	collectorContext := r.Context().Value(collectorContextKey).(collectorContext)

	// status counter
	s := collectorContext.ResponseWriter.Status()
	key := MetricsInstance().GetProcessedTotalMapKey(metricPathKey, s, requestHost)
	c, ok := MetricsInstance().Get(key)
	if ok {
		c.(prometheus.Counter).Inc()
	}

	// request latency
	totalLatency := time.Since(collectorContext.StartTime).Seconds()
	key = MetricsInstance().GetRequestLatencyMapKey(metricPathKey, requestHost)
	c, ok = MetricsInstance().Get(key)
	if ok {
		c.(prometheus.Observer).Observe(totalLatency)
	}

	proxyContext := chainContext.Proxy
	// upstream request latency
	upstreamLatency := time.Since(proxyContext.UpstreamRequestStartTime).Seconds()

	key = MetricsInstance().GetUpstreamRequestLatencyMapKey(metricPathKey, requestHost)
	c, ok = MetricsInstance().Get(key)
	if ok {
		c.(prometheus.Observer).Observe(upstreamLatency)
	}

	if chainContext.Conf.Middlewares.Cache.IsEnabled() {
		cacheContext := chainContext.Cache
		key = MetricsInstance().GetCacheTotalMapKey(metricPathKey, cacheContext.Status, requestHost)
		c, ok = MetricsInstance().Get(key)
		if ok {
			c.(prometheus.Counter).Inc()
		}
	}
}

func (m *EmitterMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	m.emitLogs(r)
	m.emitMetrics(r)
	m.next.ServeHTTP(w, r)
}
