package timeout

import (
	"net/http"

	loggerutils "github.com/ferama/crauti/pkg/logger/utils"
	"github.com/ferama/crauti/pkg/middleware"
)

type timeoutHandlerMiddleware struct {
	middleware.Middleware

	next http.Handler
}

func NewTimeoutHandlerMiddleware(next http.Handler) *timeoutHandlerMiddleware {
	m := &timeoutHandlerMiddleware{
		next: next,
	}
	return m
}

func (m *timeoutHandlerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	select {
	// a timeout occurred?
	case <-r.Context().Done():
		w.Write([]byte("bad gateway: connection timeout\n"))
		loggerutils.EmitAndReturn(w, r)
		return
	default:
	}

	m.next.ServeHTTP(w, r)
}
