package timeout

import (
	"context"
	"net/http"
	"time"
)

type TimeoutMiddleware struct {
	next    http.Handler
	timeout time.Duration
}

func NewTimeoutMiddleware(next http.Handler, timeout time.Duration) http.Handler {
	m := &TimeoutMiddleware{
		next:    next,
		timeout: timeout,
	}
	return m
}

func (m *TimeoutMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), m.timeout)
	r = r.WithContext(ctx)

	defer func() {
		cancel()
		// w.WriteHeader(http.StatusGatewayTimeout)
		// w.Write([]byte("bad gateway: connection timeout\n"))

	}()

	m.next.ServeHTTP(w, r)
}
