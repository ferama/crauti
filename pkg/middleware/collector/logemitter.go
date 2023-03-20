package collector

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/ferama/crauti/pkg/middleware/cache"
	"github.com/ferama/crauti/pkg/middleware/proxy"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type logEmitterMiddleware struct {
	next http.Handler
}

func NewLogEmitterrMiddleware(next http.Handler) http.Handler {
	m := &logEmitterMiddleware{
		next: next,
	}
	return m
}

func (m *logEmitterMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
			Float64("latency", upstreamLatency)

		event.Dict("proxyUpstream", proxyUpstreamDict)

	}

	event.Send()

	m.next.ServeHTTP(w, r)
}
