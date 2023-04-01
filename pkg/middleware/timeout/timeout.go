package timeout

import (
	"context"
	"net/http"

	"github.com/ferama/crauti/pkg/middleware"
)

type timeoutMiddleware struct {
	middleware.Middleware

	next http.Handler
}

func NewTimeoutMiddleware(next http.Handler) *timeoutMiddleware {
	m := &timeoutMiddleware{
		next: next,
	}
	return m
}

func (m *timeoutMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chainContext := m.GetContext(r)

	timeout := chainContext.Conf.Middlewares.Timeout

	if timeout > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		r = r.WithContext(ctx)

		defer cancel()
	}

	m.next.ServeHTTP(w, r)
}
