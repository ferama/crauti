package logger

import (
	"fmt"
	"net/http"
	"time"

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
	lc := r.Context().Value(loggerContextKey).(logCollectorContext)
	ww := lc.ResponseWriter

	elapsed := time.Since(lc.StartTime).Round(1 * time.Millisecond).Seconds()
	log.Info().
		Int("status", ww.Status()).
		Int("responseSize", ww.BytesWritten()).
		Float64("latency", elapsed).
		Str("userAgent", r.UserAgent()).
		Str("url", fmt.Sprintf("%s%s", r.Host, r.URL.Path)).
		Send()

	m.next.ServeHTTP(w, r)
}
