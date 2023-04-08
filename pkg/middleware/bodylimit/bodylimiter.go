package bodylimit

import (
	"net/http"
	"sync"

	"github.com/ferama/crauti/pkg/chaincontext"
	"github.com/ferama/crauti/pkg/middleware"
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

type BodyLimiterMiddleware struct {
	middleware.Middleware

	next http.Handler
}

func (m *BodyLimiterMiddleware) Init(next http.Handler) middleware.Middleware {
	m.next = next
	return m
}

func (m *BodyLimiterMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chainContext := chaincontext.GetChainContext(r)
	maxSize, _ := utils.ConvertToBytes(chainContext.Conf.Middlewares.MaxRequestBodySize)

	// unlimited
	if maxSize <= 0 {
		m.next.ServeHTTP(w, r)
		return
	}

	if r.ContentLength > maxSize {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reader := limiterPool.Get().(*limiterReader)
	defer limiterPool.Put(reader)

	reader.Reset(r.Body, maxSize)
	r.Body = reader

	m.next.ServeHTTP(w, r)
}
