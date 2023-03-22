package collector

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ferama/crauti/pkg/middleware/cache"
	"github.com/ferama/crauti/pkg/middleware/proxy"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type logEmitterMiddleware struct {
	next http.Handler

	metricPrefix string
}

func NewLogEmitterrMiddleware(next http.Handler, mountPointPath string) http.Handler {
	// TODO: build metric prefix
	// each mount point should hold its own metrics
	// then I could have global one too
	// Maintaintain a prometheus.Collector map and cleanup the collectors
	// on server.UpdateHandlers
	mp := strings.ToLower(mountPointPath)

	// prometheus.DefaultRegisterer.Unregister(crautiOpsProcessed)

	m := &logEmitterMiddleware{
		next:         next,
		metricPrefix: mp,
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
	crautiOpsProcessed.Inc()
}

func (m *logEmitterMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	m.emitLogs(r)
	m.emitMetrics(r)
	m.next.ServeHTTP(w, r)
}
