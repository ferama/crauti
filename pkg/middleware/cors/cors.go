package cors

import (
	"log"
	"net/http"
)

// this is a simple middleware sample
type corsMiddleware struct {
	next http.Handler
}

func NewCorsMiddleware(next http.Handler) http.Handler {
	h := &corsMiddleware{
		next: next,
	}
	return h
}

func (m *corsMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("test 2")

	rw := &responseWriter{
		w: w,
		r: r,
	}
	m.next.ServeHTTP(rw, r)
}
