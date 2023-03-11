package cors

import (
	"net/http"
)

// this is a simple middleware sample
type Cors struct {
	next http.Handler
}

func NewCors(next http.Handler) http.Handler {
	h := &Cors{
		next: next,
	}
	return h
}

func (h *Cors) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rw := &responseWriter{
		w: w,
		r: r,
	}
	h.next.ServeHTTP(rw, r)
}
