package collector

import (
	"context"
	"net/http"
	"sync"
	"time"

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

type contextKey string

const collectorContextKey contextKey = "collector-middleware-context"

type collectorContext struct {
	ResponseWriter *responseWriter
	StartTime      time.Time
}

// The collector middleare, lives for the entire request duration. It needs
// to be the first middleware executed. It will collect all sort of metrics
// and request related stuff like response status and stuff.
type CollectorMiddleware struct {
	middleware.Middleware

	next http.Handler
}

func (m *CollectorMiddleware) Init(next http.Handler) middleware.Middleware {
	m.next = next
	return m
}

func (m *CollectorMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// use a custom response writer to be able to capture stuff
	// like status and response bytes written...
	rw := responseWriterPool.Get().(*responseWriter)
	defer responseWriterPool.Put(rw)
	rw.Reset(r, w)

	// ... and put the response writer into the context to be accessed
	// from emitter
	lcc := collectorContext{
		ResponseWriter: rw,
		StartTime:      time.Now(),
	}
	ctx := context.WithValue(r.Context(), collectorContextKey, lcc)
	r = r.WithContext(ctx)

	m.next.ServeHTTP(rw, r)

	// the emitting stage is handled by logemitter. I cannot do
	// it here becouse at this point I don't have the full context required for
	// logging (think about the cache context for example)
}
