package cors

import (
	"net/http"
	"sync"

	"github.com/ferama/crauti/pkg/middleware"
)

var responseWriterPool sync.Pool

func init() {
	responseWriterPool = sync.Pool{
		New: func() any {
			r := &responseWriter{}
			return r
		},
	}
}

// this is a simple middleware sample
type corsMiddleware struct {
	middleware.Middleware

	next http.Handler
}

func NewCorsMiddleware(next http.Handler) *corsMiddleware {
	h := &corsMiddleware{
		next: next,
	}
	return h
}

func (m *corsMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chainContext := m.GetChainContext(r)

	if chainContext.Conf.Middlewares.Cors.IsEnabled() {

		rw := responseWriterPool.Get().(*responseWriter)
		defer responseWriterPool.Put(rw)
		rw.Reset(r, w)

		m.next.ServeHTTP(rw, r)
	} else {
		m.next.ServeHTTP(w, r)
	}
}
