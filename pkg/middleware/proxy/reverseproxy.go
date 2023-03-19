package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/logger"
	"github.com/ferama/crauti/pkg/middleware/cache"
	"github.com/rs/zerolog"
)

var log *zerolog.Logger

func init() {
	log = logger.GetLogger("reverseproxy")
}

type contextKey string

const ProxyContextKey contextKey = "proxy-middleware-context"

type ProxyContext struct {
	Upstream *url.URL
}

type reverseProxyMiddleware struct {
	next http.Handler
	// the upstream url
	upstream *url.URL
	// the request directet to this mountPath will be proxied to the upstream
	mountPath string
	// a reverse proxy instance
	rp *httputil.ReverseProxy
}

func NewReverseProxyMiddleware(
	next http.Handler,
	mountPoint conf.MountPoint,
) http.Handler {

	upstreamUrl, err := url.Parse(mountPoint.Upstream)
	if err != nil {
		log.Fatal().Err(err)
	}

	p := &reverseProxyMiddleware{
		next:      next,
		upstream:  upstreamUrl,
		rp:        httputil.NewSingleHostReverseProxy(upstreamUrl),
		mountPath: mountPoint.Path,
	}

	director := p.rp.Director
	p.rp.Director = func(r *http.Request) {
		director(r)
		// set the request host to the real upstream host
		// When is this really needed? Enabling this one
		// causes not wanted behaviours trying to fully prox a virtual host
		// target
		// r.Host = upstreamUrl.Host

		// This to support configs like:
		// - upstream: https://api.myurl.cloud/config/v1/apps
		//	 path: /api/config/v1/apps
		// This allow to fine tune proxy config for each upstream endpoint
		if !strings.HasSuffix(p.mountPath, "/") {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}
	}
	p.rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Debug().
			Str("upstream", fmt.Sprintf("%s://%s:%s", p.upstream.Scheme, p.upstream.Host, p.upstream.Port())).
			Msg(err.Error())
	}

	p.rp.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return p
}

func (m *reverseProxyMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	r = r.WithContext(context.WithValue(
		r.Context(),
		ProxyContextKey,
		ProxyContext{Upstream: m.upstream}))

	cacheContext := r.Context().Value(cache.CacheContextKey)
	// if we do not have tha cache middleware enabled or if it is enabled but the requests
	// doesn't hit the cache, poke the upstream
	if cacheContext == nil || cacheContext.(cache.CacheContext).Status != cache.CacheStatusHit {
		h := http.StripPrefix(m.mountPath, m.rp)
		h.ServeHTTP(w, r)
	}
	m.next.ServeHTTP(w, r)
}
