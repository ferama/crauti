package reverseproxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/ferama/crauti/pkg/conf"
)

type ReverseProxy struct {
	next http.Handler
	// the upstream url
	upstream *url.URL
	// the request directet to this mountPath will be proxied to the upstream
	mountPath string
	// a reverse proxy instance
	rp *httputil.ReverseProxy
}

func NewReverseProxy(
	next http.Handler,
	mountPoint *conf.MountPoint,
) (http.Handler, error) {

	upstreamUrl, err := url.Parse(mountPoint.Upstream)
	if err != nil {
		return nil, err
	}

	p := &ReverseProxy{
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

	return p, nil
}

func (p *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := http.StripPrefix(p.mountPath, p.rp)
	h.ServeHTTP(w, r)
	p.next.ServeHTTP(w, r)
}
