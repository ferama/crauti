package cors

import (
	"net/http"
	"sync"

	"github.com/ferama/crauti/pkg/chaincontext"
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
type CorsMiddleware struct {
	middleware.Middleware

	next http.Handler
}

func (m *CorsMiddleware) Init(next http.Handler) middleware.Middleware {
	m.next = next
	return m
}

func (m *CorsMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chainContext := chaincontext.GetChainContext(r)

	if chainContext.Conf.Middlewares.Cors.IsEnabled() {

		rw := responseWriterPool.Get().(*responseWriter)
		defer responseWriterPool.Put(rw)
		rw.Reset(r, w)

		m.next.ServeHTTP(rw, r)
	} else {
		m.next.ServeHTTP(w, r)
	}
}
