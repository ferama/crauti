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
		// root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		// var chain http.Handler
		// chain = root

		// chain = loggermiddleware.NewLogPrinterMiddleware(chain)
		// chain.ServeHTTP(w, r)
	}()

	r = r.WithContext(ctx)
	m.next.ServeHTTP(w, r)
}
