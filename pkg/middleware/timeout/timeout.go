package timeout

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
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

		// BUG TODO: logs printer middleware do not get the Timeout Status
		w.WriteHeader(http.StatusGatewayTimeout)
		w.Write([]byte("bad gateway: connection timeout\n"))
		log.Info().Msg("test")
	}()

	r = r.WithContext(ctx)
	m.next.ServeHTTP(w, r)
}
