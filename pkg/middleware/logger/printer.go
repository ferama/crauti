package logger

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ferama/crauti/pkg/middleware/cache"
	"github.com/rs/zerolog/log"
)

type logPrinterMiddleware struct {
	next http.Handler
}

func NewLogPrinterMiddleware(next http.Handler) http.Handler {
	m := &logPrinterMiddleware{
		next: next,
	}
	return m
}

func (m *logPrinterMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logContext := r.Context().Value(loggerContextKey).(logCollectorContext)

	elapsed := time.Since(logContext.StartTime).Round(1 * time.Millisecond).Seconds()
	event := log.Info().
		Int("status", logContext.ResponseWriter.Status()).
		Int("responseSize", logContext.ResponseWriter.BytesWritten()).
		Float64("latency", elapsed).
		Str("userAgent", r.UserAgent()).
		Str("url", fmt.Sprintf("%s%s", r.Host, r.URL.Path))

	cacheContext := r.Context().Value(cache.CacheContextKey)
	if cacheContext != nil {
		event.Str("cache", cacheContext.(cache.CacheContext).Status)
	}
	event.Send()

	m.next.ServeHTTP(w, r)
}
