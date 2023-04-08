package timeout

import (
	"net/http"

	"github.com/ferama/crauti/pkg/middleware"
)

type TimeoutHandlerMiddleware struct {
	middleware.Middleware

	next http.Handler
}

func (m *TimeoutHandlerMiddleware) Init(next http.Handler) middleware.Middleware {
	m.next = next
	return m
}

func (m *TimeoutHandlerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	select {
	// a timeout occurred?
	case <-r.Context().Done():
		w.Write([]byte("bad gateway: connection timeout\n"))
		return
	default:
	}

	m.next.ServeHTTP(w, r)
}
