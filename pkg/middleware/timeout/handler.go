package timeout

import (
	"net/http"

	"github.com/ferama/crauti/pkg/middleware"
	"github.com/ferama/crauti/pkg/utils"
)

type timeoutHandlerMiddleware struct {
	middleware.Middleware

	next http.Handler
}

func NewTimeoutHandlerMiddleware(next http.Handler) http.Handler {
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
		utils.EmitAndReturn(w, r)
		return
	default:
	}

	m.next.ServeHTTP(w, r)
}
