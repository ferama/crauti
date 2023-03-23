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

	// still serve the next op. If a timeout occurred, next ops can always detect
	// it using the r.Context().Done() channel.
	// Actually the ReverseProxyMiddleware already handle it using the behaviour
	// inherited from httputil.NewSingleHostReverseProxy
	// Ideally the only next op should be the log emitter and no more changes should
	// be made to the response.
	m.next.ServeHTTP(w, r)
}
