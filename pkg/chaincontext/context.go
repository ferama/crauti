package chaincontext

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/ferama/crauti/pkg/conf"
)

type contextKey string

const ChainContextKey contextKey = "chain-context"

// The chain context holds all the mountPoints related config
// It is easily accessed from all the middleware without requiring
// any custom variable passing and stuff
type ChainContext struct {
	Conf  conf.MountPoint
	Proxy *ProxyContext
	Cache *CacheContext
}

func (c *ChainContext) Update(r *http.Request, n ChainContext) *http.Request {
	r = r.WithContext(context.WithValue(
		r.Context(),
		ChainContextKey,
		n,
	))
	return r
}

type ProxyContext struct {
	Upstream                 *url.URL
	MountPath                string
	UpstreamRequestStartTime time.Time
}

type CacheContext struct {
	Status string
}
