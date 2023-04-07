package redirect

import (
	"net/http"
	"strings"

	"github.com/ferama/crauti/pkg/chaincontext"
	"github.com/ferama/crauti/pkg/middleware"
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
	ctx := chaincontext.GetChainContext(r)
	if !ctx.Conf.Middlewares.IsRedirectToHTTPS() {
		m.next.ServeHTTP(w, r)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	uri := r.RequestURI

	if scheme != "https" {
		// allow acme-challenge on http
		if !strings.HasPrefix(r.URL.Path, "/.well-known/acme-challenge/") {
			url := "https://" + host + uri
			w.Header().Set("Location", url)
			w.WriteHeader(http.StatusPermanentRedirect)
			return
		}
	}
	m.next.ServeHTTP(w, r)
}
