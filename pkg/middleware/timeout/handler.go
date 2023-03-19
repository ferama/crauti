package timeout

import (
	"net/http"
)

type timeoutHandlerMiddleware struct {
	next http.Handler
}

func NewTimeoutHandlerMiddleware(next http.Handler) http.Handler {
	m := &timeoutHandlerMiddleware{
		next: next,
	}
	return m
}

func (m *timeoutHandlerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	select {
	// a timeout occurred?
	case <-r.Context().Done():
		w.WriteHeader(http.StatusGatewayTimeout)
		w.Write([]byte("bad gateway: connection timeout\n"))
	default:
	}

	m.next.ServeHTTP(w, r)
}
