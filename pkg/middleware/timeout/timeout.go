package timeout

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
	r = r.WithContext(ctx)

	defer cancel()

	m.next.ServeHTTP(w, r)
}
