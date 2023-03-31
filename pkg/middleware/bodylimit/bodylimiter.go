package bodylimit

import (
	"net/http"
	"sync"

	"github.com/ferama/crauti/pkg/middleware"
	"github.com/ferama/crauti/pkg/utils"
)

var lPool sync.Pool

func init() {
	lPool = sync.Pool{
		New: func() any {
			r := &limiterReader{}
			return r
		},
	}
}

type bodyLimiter struct {
	middleware.Middleware

	next http.Handler
}

func NewBodyLimiterMiddleware(next http.Handler) *bodyLimiter {
	m := &bodyLimiter{
		next: next,
	}
	return m
}

func (m *bodyLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chainContext := m.GetChainContext(r)
	maxSize := chainContext.Conf.Middlewares.MaxRequestBodySize

	// unlimited
	if maxSize < 0 {
		m.next.ServeHTTP(w, r)
		return
	}

	if r.ContentLength > maxSize {
		w.WriteHeader(http.StatusBadRequest)
		utils.EmitAndReturn(w, r)
		return
	}

	reader := lPool.Get().(*limiterReader)
	reader.Reset(r.Body, maxSize)
	r.Body = reader

	m.next.ServeHTTP(w, r)
}
