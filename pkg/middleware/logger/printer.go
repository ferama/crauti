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

// {
// 	"timestamp":"2023-03-18T20:30:38+00:00",
// 	"requestID":"0f9c2aecee8c1deae250e017d4f4d960",
// 	"proxyUpstreamName":"elk-elk-kibana-http",
// 	"proxyAlternativeUpstreamName":"",
// 	"upstreamStatus":"200",
// 	"upstreamAddr":"172.20.36.212:5601",
// 	"httpRequest":{
// 	   "requestMethod":"POST",
// 	   "requestUrl":"kibana.paas.relatech.link/api/lens/stats",
// 	   "status":200,
// 	   "requestSize":"96",
// 	   "responseSize":"2",
// 	   "userAgent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36",
// 	   "remoteIp":"10.99.99.2",
// 	   "referer":"https://kibana.paas.relatech.link/app/lens",
// 	   "latency":"1.007 s",
// 	   "protocol":"HTTP/2.0"
// 	}
//  }

func Printer(w http.ResponseWriter, r *http.Request) {
	logContext := r.Context().Value(loggerContextKey).(logCollectorContext)

	elapsed := time.Since(logContext.StartTime).Round(1 * time.Millisecond).Seconds()

	url := fmt.Sprintf("%s%s?%s", r.Host, r.URL.Path, r.URL.RawQuery)
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
}
