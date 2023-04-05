package redirect

import (
	"net/http"

	"github.com/ferama/crauti/pkg/middleware"
	"github.com/rs/zerolog/log"
)

type RedirectMiddleware struct {
	middleware.Middleware

	next http.Handler
}

func (m *RedirectMiddleware) Init(next http.Handler) middleware.Middleware {
	m.next = next
	return m
}

func (m *RedirectMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	uri := r.RequestURI

	log.Printf("s: %s, h: %s, u: %s", scheme, host, uri)
	if scheme != "https" {
		url := "https://" + host + uri
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusPermanentRedirect)
		return
	}
	m.next.ServeHTTP(w, r)
}
