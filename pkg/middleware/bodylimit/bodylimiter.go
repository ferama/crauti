package bodylimit

import (
	"net/http"
	"sync"

	"github.com/ferama/crauti/pkg/middleware"
	collectorutils "github.com/ferama/crauti/pkg/middleware/collector/utils"
	"github.com/ferama/crauti/pkg/utils"
)

var limiterPool sync.Pool

func init() {
	limiterPool = sync.Pool{
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
	chainContext := m.GetContext(r)
	maxSize, _ := utils.ConvertToBytes(chainContext.Conf.Middlewares.MaxRequestBodySize)

	// unlimited
	if maxSize <= 0 {
		m.next.ServeHTTP(w, r)
		return
	}

	if r.ContentLength > maxSize {
		w.WriteHeader(http.StatusBadRequest)
		collectorutils.EmitAndReturn(w, r)
		return
	}

	reader := limiterPool.Get().(*limiterReader)
	defer limiterPool.Put(reader)

	reader.Reset(r.Body, maxSize)
	r.Body = reader

	m.next.ServeHTTP(w, r)
}
