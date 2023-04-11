package collector

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/ferama/crauti/pkg/chaincontext"
	"github.com/ferama/crauti/pkg/middleware"
	"github.com/ferama/crauti/pkg/utils"
)

var responseWriterPool sync.Pool

func init() {
	responseWriterPool = sync.Pool{
		New: func() any {
			r := &responseWriter{}
			return r
		},
	}
}

type contextKey string

const collectorContextKey contextKey = "collector-middleware-context"

type collectorContext struct {
	ResponseWriter *responseWriter
	StartTime      time.Time
}

// The collector middleare, lives for the entire request duration. It needs
// to be the first middleware executed. It will collect all sort of metrics
// and request related stuff like response status and stuff.
type CollectorMiddleware struct {
	middleware.Middleware

	next http.Handler
}

func (m *CollectorMiddleware) Init(next http.Handler) middleware.Middleware {
	m.next = next
	return m
}

func (m *CollectorMiddleware) emitLogs(r *http.Request) {
	ctx := chaincontext.GetChainContext(r)

	collectorContext := r.Context().Value(collectorContextKey).(collectorContext)

	totalLatency := time.Since(collectorContext.StartTime)

	uri := utils.GetURI(r.URL)

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
		Str("latencyHuman", totalLatency.Round(1*time.Millisecond).String()).
		Str("protocol", r.Proto)

	event := log.Info().
		Dict("httpRequest", httpRequestDict)

	if ctx.Conf.Middlewares.Cache.IsEnabled() {
		cacheContext := ctx.Cache
		event.Str("cache", cacheContext.Status)
	}

	proxyContext := ctx.Proxy
	upstreamLatency := time.Since(proxyContext.UpstreamRequestStartTime)

	proxyUpstreamDict := zerolog.Dict().
		Str("url", ctx.Conf.Upstream).
		Str("mountPath", ctx.Conf.Path).
		Str("uri", ctx.Proxy.URI).
		Float64("latency", upstreamLatency.Seconds()).
		Str("latencyHuman", upstreamLatency.Round(1*time.Millisecond).String())

	event.Dict("proxyUpstream", proxyUpstreamDict)

	event.Send()
}

func (m *CollectorMiddleware) emitMetrics(r *http.Request) {
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

func (m *CollectorMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	defer func() {
		m.emitLogs(r)
		m.emitMetrics(r)
	}()

	// use a custom response writer to be able to capture stuff
	// like status and response bytes written...
	rw := responseWriterPool.Get().(*responseWriter)
	defer responseWriterPool.Put(rw)
	rw.Reset(r, w)

	// ... and put the response writer into the context to be accessed
	// from emitter
	lcc := collectorContext{
		ResponseWriter: rw,
		StartTime:      time.Now(),
	}
	ctx := context.WithValue(r.Context(), collectorContextKey, lcc)
	r = r.WithContext(ctx)

	m.next.ServeHTTP(rw, r)
}
