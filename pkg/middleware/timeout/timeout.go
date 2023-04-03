package timeout

import (
	"context"
	"net/http"

	"github.com/ferama/crauti/pkg/chaincontext"
	"github.com/ferama/crauti/pkg/middleware"
)

type TimeoutMiddleware struct {
	middleware.Middleware

	next http.Handler
}

func (m *TimeoutMiddleware) Init(next http.Handler) middleware.Middleware {
	m.next = next
	return m
}
func (m *TimeoutMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chainContext := chaincontext.GetChainContext(r)

	timeout := chainContext.Conf.Middlewares.Timeout

	if timeout > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		r = r.WithContext(ctx)

		defer cancel()
	}

	m.next.ServeHTTP(w, r)
}
