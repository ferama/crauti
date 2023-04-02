package proxy

import (
	"crypto/tls"
	"fmt"
	"io"
	golog "log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/logger"
	"github.com/ferama/crauti/pkg/middleware"
	"github.com/ferama/crauti/pkg/middleware/cache"
	collectorutils "github.com/ferama/crauti/pkg/middleware/collector/utils"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

var log *zerolog.Logger

var bpool *bufferPool

func init() {
	// this one is here to make some init vars available to other
	// init functions.
	// The use case is the CRAUTI_DEBUG that need to be available as
	// soon as possibile in order to instantiate the logger correctly
	viper.ReadInConfig()
	conf.Update()
	// apparently there is no way to use a custom logger like zerolog
	// Most cases are handled overriding the ErrorHandler
	// Disabling golog here, it should only affect the copy buffer failure that is already
	// handled recovering from panic in the ServeHTTP method below
	golog.SetOutput(io.Discard)

	log = logger.GetLogger("reverseproxy")
	bpool = newPool()
}

type reverseProxyMiddleware struct {
	middleware.Middleware

	next http.Handler
}

func NewReverseProxyMiddleware(next http.Handler) *reverseProxyMiddleware {

	p := &reverseProxyMiddleware{
		next: next,
	}
	return p
}

func (m *reverseProxyMiddleware) director(proxy *httputil.ReverseProxy) func(r *http.Request) {
	director := proxy.Director

	return func(r *http.Request) {
		director(r)

		chainContext := m.GetContext(r)
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

	// install the buffer pool
	proxy.BufferPool = bpool
	proxy.Director = m.director(proxy)

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Debug().
			Str("upstream", fmt.Sprintf("%s://%s", upstreamUrl.Scheme, upstreamUrl.Host)).
			Msg(err.Error())

		select {
		case <-r.Context().Done():
			w.WriteHeader(http.StatusGatewayTimeout)
		default:
			w.WriteHeader(http.StatusBadGateway)
		}
	}
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return proxy
}

func (m *reverseProxyMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := m.GetContext(r)

	upstreamUrl, err := url.Parse(ctx.Conf.Upstream)
	if err != nil {
		log.Fatal().Err(err)
	}

	matchHost := ctx.Conf.Middlewares.Proxy.MatchHost
	// TODO: it always work?
	requestHost, _, err := net.SplitHostPort(r.Host)

	if err != nil {
		log.Fatal().Err(err)
	}
	if matchHost != "" {
		if matchHost != requestHost {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "crauti: 404 not found\n")
			collectorutils.EmitAndReturn(w, r)
			return
		}
	}

	proxy := m.setupProxy(upstreamUrl)

	ctx.Proxy.UpstreamRequestStartTime = time.Now()

	r = ctx.Update(r, ctx)

	cacheContext := ctx.Cache
	// if we do not have tha cache middleware enabled or if it is enabled but the requests
	// doesn't hit the cache, poke the upstream
	cacheEnabled := ctx.Conf.Middlewares.Cache.IsEnabled()
	if !cacheEnabled || cacheContext.Status != cache.CacheStatusHit {
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
		ctx.Proxy.ProxiedRequest = true
		proxy.ServeHTTP(w, r)

	} else {
		log.Debug().
			Str("upstream", fmt.Sprintf("%s://%s", upstreamUrl.Scheme, upstreamUrl.Host)).
			Msg("do not poke upstream: already got from cache")
	}
	m.next.ServeHTTP(w, r)
}
