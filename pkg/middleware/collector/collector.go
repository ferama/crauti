package collector

import (
	"context"
	"net/http"
	"time"
)

type contextKey string

const collectorContextKey contextKey = "collector-middleware-context"

type collectorContext struct {
	ResponseWriter *responseWriter
	StartTime      time.Time
}

// The collector middleare, lives for the entire request duration. It needs
// to be the first middleware executed. It will collect all sort of metrics
// and request related stuff like response status and stuff.
type collectorMiddleware struct {
	next http.Handler
}

func NewCollectorMiddleware(next http.Handler) http.Handler {
	m := &collectorMiddleware{
		next: next,
	}
	return m
}

func (m *collectorMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// use a custom response writer to be able to capture stuff
	// like status and response bytes written...
	rw := &responseWriter{
		w: w,
		r: r,
	}

	// ... and put the response writer into the context to be accessed
	// from emitter
	lcc := collectorContext{
		ResponseWriter: rw,
		StartTime:      time.Now(),
	}
	ctx := context.WithValue(r.Context(), collectorContextKey, lcc)
	r = r.WithContext(ctx)

	m.next.ServeHTTP(rw, r)
}
