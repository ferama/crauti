package logger

import (
	"context"
	"net/http"
	"time"
)

type contextKey string

const loggerContextKey contextKey = "logcollector-middleware-context"

type logCollectorContext struct {
	ResponseWriter WrapResponseWriter
	StartTime      time.Time
}

type logCollectorMiddleware struct {
	next http.Handler
}

func NewLogCollectorMiddleware(next http.Handler) http.Handler {
	m := &logCollectorMiddleware{
		next: next,
	}
	return m
}

func (m *logCollectorMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ww := NewWrapResponseWriter(w, r.ProtoMajor)

	lcc := logCollectorContext{
		ResponseWriter: ww,
		StartTime:      time.Now(),
	}
	ctx := context.WithValue(r.Context(), loggerContextKey, lcc)
	r = r.WithContext(ctx)
	m.next.ServeHTTP(ww, r)
}
