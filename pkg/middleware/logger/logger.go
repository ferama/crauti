package logger

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ferama/crauti/pkg/logger"
	"github.com/rs/zerolog"
)

var log *zerolog.Logger

func init() {
	log = logger.GetLogger("logger")
}

type loggerMiddleware struct {
	next http.Handler
}

func NewLoggerMiddleware(next http.Handler) http.Handler {
	m := &loggerMiddleware{
		next: next,
	}
	return m
}

func (m *loggerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ww := NewWrapResponseWriter(w, r.ProtoMajor)
	t1 := time.Now()
	defer func() {
		elapsed := time.Since(t1).Round(1 * time.Millisecond).Seconds()
		log.Info().
			Int("status", ww.Status()).
			Int("responseSize", ww.BytesWritten()).
			Float64("latency", elapsed).
			Str("userAgent", r.UserAgent()).
			Str("url", fmt.Sprintf("%s%s", r.Host, r.URL.Path)).
			Send()
	}()

	m.next.ServeHTTP(ww, r)
}
