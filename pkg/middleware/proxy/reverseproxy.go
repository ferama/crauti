package proxy

import (
	"crypto/tls"
	"fmt"
	"io"
	golog "log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/ferama/crauti/pkg/chaincontext"
	"github.com/ferama/crauti/pkg/logger"
	"github.com/ferama/crauti/pkg/middleware"
	"github.com/ferama/crauti/pkg/middleware/cache"
	"github.com/rs/zerolog"
)

var log *zerolog.Logger

func init() {
	// apparently there is no way to use a custom logger like zerolog
	// Most cases are handled overriding the ErrorHandler
	// Disabling golog here, it should only affect the copy buffer failure that is already
	// handled recovering from panic in the ServeHTTP method below
	golog.SetOutput(io.Discard)

	log = logger.GetLogger("reverseproxy")
}

type reverseProxyMiddleware struct {
	middleware.Middleware

	next http.Handler
	// the upstream url
}

func NewReverseProxyMiddleware(next http.Handler) http.Handler {

	p := &reverseProxyMiddleware{
		next: next,
	}
	return p
}

func (m *reverseProxyMiddleware) director(proxy *httputil.ReverseProxy) func(r *http.Request) {
	director := proxy.Director

	return func(r *http.Request) {
		director(r)

		chainContext := m.GetChainContext(r)
		upstreamUrl, err := url.Parse(chainContext.Conf.Upstream)
		if err != nil {
			log.Fatal().Err(err)
		}
		// set the request host to the real upstream host
		if chainContext.Conf.Middlewares.Proxy.IsHostHeaderPreserved() {
			r.Host = upstreamUrl.Host
		}

		// This to support configs like:
		// - upstream: https://api.myurl.cloud/config/v1/apps
		//	 path: /api/config/v1/apps
		// This allow to fine tune proxy config for each upstream endpoint
		if !strings.HasSuffix(chainContext.Conf.Path, "/") {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}
	}
}

func (m *reverseProxyMiddleware) setupProxy(upstreamUrl *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(upstreamUrl)
	proxy.Director = m.director(proxy)

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Debug().
			Str("upstream", fmt.Sprintf("%s://%s", upstreamUrl.Scheme, upstreamUrl.Host)).
			Msg(err.Error())
		w.WriteHeader(http.StatusBadGateway)
	}
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return proxy
}

func (m *reverseProxyMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := m.GetChainContext(r)

	upstreamUrl, err := url.Parse(ctx.Conf.Upstream)
	if err != nil {
		log.Fatal().Err(err)
	}

	proxy := m.setupProxy(upstreamUrl)

	ctx.Proxy = &chaincontext.ProxyContext{
		UpstreamRequestStartTime: time.Now(),
	}
	r = ctx.Update(r, ctx)

	cacheContext := ctx.Cache
	// if we do not have tha cache middleware enabled or if it is enabled but the requests
	// doesn't hit the cache, poke the upstream
	if cacheContext == nil || cacheContext.Status != cache.CacheStatusHit {
		log.Debug().
			Str("upstream", fmt.Sprintf("%s://%s", upstreamUrl.Scheme, upstreamUrl.Host)).
			Msg("poke upstream")

		proxy := http.StripPrefix(ctx.Conf.Path, proxy)

		defer func() {
			// the call to proxy.ServeHTTP some rows below, will panic if
			// the request is aborted client side. The panic is transparent (it is handled
			// somewhere, needs investigation). The point is that an aborted request
			// is not logged anywhere and this code is needed just to do that.
			if rec := recover(); rec != nil {
				log.Error().
					Str("upstream", fmt.Sprintf("%s://%s", upstreamUrl.Scheme, upstreamUrl.Host)).
					Msg("request aborted")

				// Even if the request is aborted I'm processing the next chain ring
				// here that actually is the timeoutHandler followed by the log emitter
				m.next.ServeHTTP(w, r)
			}
		}()
		proxy.ServeHTTP(w, r)

	} else {
		log.Debug().
			Str("upstream", fmt.Sprintf("%s://%s", upstreamUrl.Scheme, upstreamUrl.Host)).
			Msg("do not poke upstream: already got from cache")
	}
	m.next.ServeHTTP(w, r)
}
