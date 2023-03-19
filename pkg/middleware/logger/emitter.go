package logger

import (
	"fmt"
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
}

func NewLogEmitterrMiddleware(next http.Handler) http.Handler {
	m := &logEmitterMiddleware{
		next: next,
	}
	return m
}

func (m *logEmitterMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logContext := r.Context().Value(loggerContextKey).(logCollectorContext)

	elapsed := time.Since(logContext.StartTime).Round(1 * time.Millisecond).Seconds()

	url := fmt.Sprintf("%s%s", r.Host, r.URL.Path)
	if r.URL.RawQuery != "" {
		url = fmt.Sprintf("%s?%s", url, r.URL.RawQuery)
	}
	remoteAddr := strings.Split(r.RemoteAddr, ":")[0]
	event := log.Info().
		Dict("httpRequest", zerolog.Dict().
			Str("requestMethod", r.Method).
			Str("requestUrl", url).
			Int("status", logContext.ResponseWriter.Status()).
			Int64("requestSize", r.ContentLength).
			Int("responseSize", logContext.ResponseWriter.BytesWritten()).
			Str("userAgent", r.UserAgent()).
			Str("remoteIp", remoteAddr).
			Str("referer", r.Referer()).
			Float64("latency", elapsed).
			Str("protocol", r.Proto),
		)

	cacheContext := r.Context().Value(cache.CacheContextKey)
	if cacheContext != nil {
		event.Str("cache", cacheContext.(cache.CacheContext).Status)
	}

	proxyContext := r.Context().Value(proxy.ProxyContextKey)
	if proxyContext != nil {
		pc := proxyContext.(proxy.ProxyContext)
		upstream := fmt.Sprintf("%s://%s", pc.Upstream.Scheme, pc.Upstream.Hostname())
		event.Str("proxyUpstream", upstream)
	}

	event.Send()

	m.next.ServeHTTP(w, r)
}
