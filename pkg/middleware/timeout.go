package middleware

import (
	"context"
	"net/http"
	"time"
)

type timeoutMiddleware struct {
	next    http.Handler
	timeout time.Duration
}

func NewTimeoutMiddleware(next http.Handler, timeout time.Duration) http.Handler {
	m := &timeoutMiddleware{
		next:    next,
		timeout: timeout,
	}
	return m
}

func (m *timeoutMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), m.timeout)
	defer func() {
		cancel()
		w.Write([]byte("bad gateway: connection timeout\n"))
	}()

	r = r.WithContext(ctx)
	m.next.ServeHTTP(w, r)
}
