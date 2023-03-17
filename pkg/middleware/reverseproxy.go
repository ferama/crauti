package middleware

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/logger"
	"github.com/rs/zerolog"
)

var log *zerolog.Logger

func init() {
	log = logger.GetLogger("reverseproxy")
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
) (http.Handler, error) {

	upstreamUrl, err := url.Parse(mountPoint.Upstream)
	if err != nil {
		return nil, err
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
		r.Host = upstreamUrl.Host
		// This to support configs like:
		// - upstream: https://api.myurl.cloud/config/v1/apps
		//	 path: /api/config/v1/apps
		// This allow to fine tune proxy config for each upstream endpoint
		// TODO: verify if we can have any side effects
		if !strings.HasSuffix(p.mountPath, "/") {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}
	}
	p.rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Error().
			Str("upstream", fmt.Sprintf("%s://%s:%s", p.upstream.Scheme, p.upstream.Host, p.upstream.Port())).
			Msg(err.Error())
	}

	return p, nil
}

func (m *reverseProxyMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := http.StripPrefix(m.mountPath, m.rp)
	h.ServeHTTP(w, r)
	m.next.ServeHTTP(w, r)
}
