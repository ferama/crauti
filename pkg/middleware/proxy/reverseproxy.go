package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

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
	Upstream                 *url.URL
	UpstreamRequestStartTime time.Time
}

type reverseProxyMiddleware struct {
	next http.Handler
	// the upstream url
	upstream *url.URL
	// the request directet to this mountPath will be proxied to the upstream
	mountPoint conf.MountPoint
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
		next:       next,
		upstream:   upstreamUrl,
		rp:         httputil.NewSingleHostReverseProxy(upstreamUrl),
		mountPoint: mountPoint,
	}
	p.rp.Director = p.director()

	p.rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Debug().
			Str("upstream", fmt.Sprintf("%s://%s", p.upstream.Scheme, p.upstream.Host)).
			Msg(err.Error())
	}

	p.rp.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return p
}

func (m *reverseProxyMiddleware) director() func(r *http.Request) {
	director := m.rp.Director
	mountPoint := m.mountPoint
	upstreamUrl := m.upstream

	return func(r *http.Request) {
		director(r)
		// set the request host to the real upstream host
		if !mountPoint.SkipHostHeader {
			r.Host = upstreamUrl.Host
		}

		// This to support configs like:
		// - upstream: https://api.myurl.cloud/config/v1/apps
		//	 path: /api/config/v1/apps
		// This allow to fine tune proxy config for each upstream endpoint
		if !strings.HasSuffix(mountPoint.Path, "/") {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}
	}
}

func (m *reverseProxyMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	r = r.WithContext(context.WithValue(
		r.Context(),
		ProxyContextKey,
		ProxyContext{
			Upstream:                 m.upstream,
			UpstreamRequestStartTime: time.Now(),
		}))

	cacheContext := r.Context().Value(cache.CacheContextKey)
	// if we do not have tha cache middleware enabled or if it is enabled but the requests
	// doesn't hit the cache, poke the upstream
	if cacheContext == nil || cacheContext.(cache.CacheContext).Status != cache.CacheStatusHit {
		log.Debug().
			Str("upstream", fmt.Sprintf("%s://%s", m.upstream.Scheme, m.upstream.Host)).
			Msg("poke upstream")

		proxy := http.StripPrefix(m.mountPoint.Path, m.rp)

		defer func() {
			if rec := recover(); rec != nil {
				log.Error().
					Str("upstream", fmt.Sprintf("%s://%s", m.upstream.Scheme, m.upstream.Host)).
					Msg("request aborted")

				m.next.ServeHTTP(w, r)
			}
		}()
		proxy.ServeHTTP(w, r)

	} else {
		log.Debug().
			Str("upstream", fmt.Sprintf("%s://%s", m.upstream.Scheme, m.upstream.Host)).
			Msg("do not poke upstream: already got from cache")
	}
	m.next.ServeHTTP(w, r)
}