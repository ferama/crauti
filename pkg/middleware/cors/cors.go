package cors

import (
	"net/http"

	"github.com/ferama/crauti/pkg/middleware"
)

// this is a simple middleware sample
type corsMiddleware struct {
	middleware.Middleware

	next http.Handler
}

func NewCorsMiddleware(next http.Handler) http.Handler {
	h := &corsMiddleware{
		next: next,
	}
	return h
}

func (m *corsMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chainContext := m.GetChainContext(r)

	if chainContext.Conf.Middlewares.Cors.IsEnabled() {
		rw := &responseWriter{
			w: w,
			r: r,
		}
		m.next.ServeHTTP(rw, r)
	} else {
		m.next.ServeHTTP(w, r)
	}
}
