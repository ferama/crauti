package timeout

import (
	"context"
	"net/http"
	"time"
)

type timeoutMiddleware struct {
	next http.Handler
}

func NewTimeoutMiddleware(next http.Handler) http.Handler {
	m := &timeoutMiddleware{
		next: next,
	}
	return m
}

func (m *timeoutMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	timeout := time.Duration(3 * time.Second)
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer func() {
		cancel()
		if ctx.Err() == context.DeadlineExceeded {
			w.WriteHeader(http.StatusGatewayTimeout)
		}
	}()

	r = r.WithContext(ctx)
	m.next.ServeHTTP(w, r)
}
