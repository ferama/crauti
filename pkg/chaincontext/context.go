package chaincontext

import (
	"context"
	"net/http"
	"time"

	"github.com/ferama/crauti/pkg/conf"
)

type contextKey string

const ChainContextKey contextKey = "chain-context"

// The chain context holds all the mountPoints related config
// It is easily accessed from all the middleware without requiring
// any custom variable passing and stuff
type ChainContext struct {
	Conf  *conf.MountPoint
	Proxy *ProxyContext
	Cache *CacheContext
}

func NewChainContext() *ChainContext {
	c := &ChainContext{
		Conf: nil,
		Proxy: &ProxyContext{
			ProxiedRequest: false,
		},
		Cache: &CacheContext{},
	}
	return c
}

// Reset the context. The context is managed using a sync.Pool and this
// method reset the instances
func (c *ChainContext) Reset(conf *conf.MountPoint, cacheStatus string) {
	c.Conf = conf
	c.Proxy.ProxiedRequest = false
	c.Cache.Status = cacheStatus
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
	// Is set to true, the request effectively reached the upstream
	// If not, it probably was served from the cache
	ProxiedRequest           bool
	UpstreamRequestStartTime time.Time
}

type CacheContext struct {
	Status string
}
