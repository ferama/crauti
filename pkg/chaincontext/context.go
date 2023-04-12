package chaincontext

import (
	"context"
	"net/http"
	"time"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/utils"
	"github.com/golang-jwt/jwt/v4"
)

type contextKey string

const chainContextKey contextKey = "chain-context"

// The chain context holds all the mountPoints related config
// It is easily accessed from all the middleware without requiring
// any custom variable passing and stuff
type ChainContext struct {
	Conf  *conf.MountPoint
	Proxy *ProxyContext
	Cache *CacheContext
	Auth  *AuthContext

	request *http.Request
}

// extracts and return chaincontext from a request
func GetChainContext(r *http.Request) ChainContext {
	chainContext := r.Context().
		Value(chainContextKey).(ChainContext)
	return chainContext
}

func NewChainContext() *ChainContext {
	c := &ChainContext{
		Conf: nil,
		Proxy: &ProxyContext{
			ProxiedRequest: false,
		},
		Cache: &CacheContext{},
		Auth: &AuthContext{
			Authorized: false,
		},
	}
	return c
}

// Reset the context. The context is managed using a sync.Pool and this
// method reset the instances
func (c *ChainContext) Reset(conf *conf.MountPoint, r *http.Request) {
	c.Conf = conf
	c.Proxy.ProxiedRequest = false
	c.Cache.Status = utils.CacheStatusMiss
	c.Auth.Authorized = false
	c.request = r
}

// returns a new request object with the updated context
// example:
//
//		// sets a new value into the context
//	 	ctx.Proxy.UpstreamRequestStartTime = time.Now()
//		// gets the updated request version
//		r = ctx.Update()
//		// propagate the context
//		next.ServeHTTP(w, r)
func (c *ChainContext) Update() *http.Request {
	r := c.request

	r = r.WithContext(context.WithValue(
		r.Context(),
		chainContextKey,
		*c,
	))
	c.request = r
	return c.request
}

type ProxyContext struct {
	// Is set to true, the request effectively reached the upstream
	// If not, it probably was served from the cache
	ProxiedRequest           bool
	UpstreamRequestStartTime time.Time
	// upstream request URI including rewrites
	URI string
}

type CacheContext struct {
	Status string
}

type AuthContext struct {
	JwtClaims  jwt.MapClaims
	Authorized bool
}
