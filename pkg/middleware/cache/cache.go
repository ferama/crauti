package cache

import (
	"bufio"
	"fmt"
	"net/http"
	"net/textproto"
	"sort"
	"strings"
	"sync"

	"github.com/ferama/crauti/pkg/chaincontext"
	"github.com/ferama/crauti/pkg/logger"
	"github.com/ferama/crauti/pkg/middleware"
	"github.com/ferama/crauti/pkg/redis"
	"github.com/rs/zerolog"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	// headers
	GeneratorHeaderKey         = "X-Generator"
	CachedContentHeaderValue   = "crauti/cache"
	UpstreamContentHeaderValue = "crauti/upstream"

	// statuses
	CacheStatusBypass  = "BYP"
	CacheStatusHit     = "HIT"
	CacheStatusIgnored = "IGN"
	CacheStatusMiss    = "MIS"

	// redis key heads. The redis key is build using the format
	//  KEYHEAD:KEYENCODING
	bodyKeyHead    = "BODY"
	headersKeyHead = "HEADERS"
	statusKeyHead  = "STATUS"
)

var log *zerolog.Logger

func init() {
	log = logger.GetLogger("cache")
}

func buildRedisKey(keyHead string, key string) string {
	return fmt.Sprintf("%s:%s", keyHead, key)
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

type cacheMiddleware struct {
	middleware.Middleware

	next http.Handler

	mu      sync.Mutex
	lockmap map[string]*sync.Mutex
}

func NewCacheMiddleware(next http.Handler) http.Handler {

	cm := &cacheMiddleware{
		next:    next,
		lockmap: make(map[string]*sync.Mutex),
	}

	return cm
}

func (m *cacheMiddleware) encodeKeyHeader(r *http.Request, enc string, k string, v string) string {
	chainContext := m.GetChainContext(r)
	conf := chainContext.Conf.Middlewares.Cache

	keyHeaders := []string{}
	c := cases.Title(language.English)
	for _, h := range conf.KeyHeaders {
		keyHeaders = append(keyHeaders, c.String(h))
	}

	if contains(keyHeaders, k) {
		enc = fmt.Sprintf("%s|%s", enc, v)
	}
	return enc
}

func (m *cacheMiddleware) buildCacheKey(r *http.Request) string {
	// ensure sorted headers (to target the right cache key)
	keys := make([]string, 0, len(r.Header))
	for k := range r.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	enc := fmt.Sprintf("%s%s", r.Method, r.URL)
	for _, k := range keys {
		v := r.Header.Get(k)
		enc = m.encodeKeyHeader(r, enc, k, v)
	}
	// claimsContext := r.Context().Value(auth.AuthContextKey)
	// if claimsContext != nil {
	// 	claims := claimsContext.(jwt.MapClaims)

	// 	keys := make([]string, 0, len(claims))
	// 	for k := range claims {
	// 		keys = append(keys, k)
	// 	}
	// 	sort.Strings(keys)
	// 	// if claim should be used to build cache key, do it!
	// 	for _, key := range keys {
	// 		val := claims[key]
	// 		if contains(p.keyClaims, key) {
	// 			enc = fmt.Sprintf("%s/%s", enc, val)
	// 		}
	// 	}
	// }

	return enc
}

func (m *cacheMiddleware) serveFromCache(key string, w http.ResponseWriter, r *http.Request) bool {
	body, _ := redis.CacheInstance().Get(buildRedisKey(bodyKeyHead, key))
	if body != nil {
		log.Debug().
			Str("status", CacheStatusHit).
			Str("key", key).Send()

		// retrieve headers string from the cache, recontsruct them
		// and put into response
		headers, _ := redis.CacheInstance().Get(buildRedisKey(headersKeyHead, key))
		if headers != nil {
			reader := bufio.NewReader(strings.NewReader(string(headers) + "\r\n"))
			tp := textproto.NewReader(reader)
			mimeHeader, err := tp.ReadMIMEHeader()
			if err != nil {
				httpHeader := http.Header(mimeHeader)
				for k, v := range httpHeader {
					w.Header().Set(k, strings.Join(v, ","))
				}
			}
		}

		w.Header().Set(GeneratorHeaderKey, CachedContentHeaderValue)

		// get cached status and write back to the response
		status, _ := redis.CacheInstance().GetInt(buildRedisKey(statusKeyHead, key))
		w.WriteHeader(status)
		w.Write(body)

		// set the hit status into the context
		chainContext := m.GetChainContext(r)
		chainContext.Cache = &chaincontext.CacheContext{
			Status: CacheStatusHit,
		}
		r = chainContext.Update(r, chainContext)

		// we can safely proceed calling the next op here. We set the cache
		// status into the context, so the next ops can adapt their behaviour using
		// this information
		m.next.ServeHTTP(w, r)
		return true
	}
	return false
}

func (m *cacheMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := m.GetChainContext(r)
	conf := ctx.Conf.Middlewares.Cache
	// if the request should not be cached because the http
	// method needs to be ignored or because it is disabled,
	// directly serve it ignoring the cache
	if !contains(conf.Methods, r.Method) || !conf.IsEnabled() {

		if conf.IsEnabled() {
			ctx.Cache = &chaincontext.CacheContext{
				Status: CacheStatusBypass,
			}
			r = ctx.Update(r, ctx)

			log.Debug().
				Str("status", CacheStatusBypass).
				Str("key", fmt.Sprintf("%s%s", r.Method, r.URL)).Send()

		}

		m.next.ServeHTTP(w, r)
		return
	}

	cacheKey := m.buildCacheKey(r)

	ignoreCache := false
	// It works like the amazon api gateway
	// https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-caching.html
	if r.Header.Get("Cache-Control") == "max-age=0" {
		log.Debug().Str("key", cacheKey).Msg("ignore cache request with Cache-Control header")
		ignoreCache = true
	}

	if !ignoreCache {
		// try to get response from cache
		if m.serveFromCache(cacheKey, w, r) {
			return
		}
		// No more then one concurrent request of the same kind (with the same enc) should hit the backend.
		// I'm using a lock here for each request kind. This will prevent multiple goroutines to
		// make same request to the backend.
		// I'm using an in memory map that is is not distributed to hold the locks (I will have one
		// per replica). The penality is that I could have 'max backend concurrent calls = max replicas'.
		//
		// The logic will prevent backend bombing. Imagine the situation where we have like 1000 concurrent
		// request at the same time to the same resource: they will all hit the cache until the cache expires.
		// If we do not put a guard here, when the cache expires, all the 1000 request will reach the
		// backend at the same time, putting a lot of pressure. The following code will prevent
		// this situations.
		var emu *sync.Mutex

		// get or create a new mutex for the cache key in a thread
		// safe way
		m.mu.Lock()
		if tmp, ok := m.lockmap[cacheKey]; ok {
			emu = tmp
		} else {
			emu = new(sync.Mutex)
			m.lockmap[cacheKey] = emu
		}
		m.mu.Unlock()

		// no more then one concurrent request to the backend for the given cache key, so
		// I'm taking the lock here
		emu.Lock()
		// cleanup the lockmap at the end and unlock
		defer func() {
			m.mu.Lock()
			defer m.mu.Unlock()

			delete(m.lockmap, cacheKey)
			emu.Unlock()
		}()

		// Another coroutine (the non locked one) likely has filled the cache already
		// so take the advantage here
		if m.serveFromCache(cacheKey, w, r) {
			return
		}
		log.Debug().
			Str("status", CacheStatusMiss).
			Str("key", cacheKey).Send()

		ctx.Cache = &chaincontext.CacheContext{
			Status: CacheStatusMiss,
		}
		r = ctx.Update(r, ctx)

	} else {
		log.Debug().
			Str("status", CacheStatusIgnored).
			Str("key", cacheKey).Send()

		ctx.Cache = &chaincontext.CacheContext{
			Status: CacheStatusIgnored,
		}
		r = ctx.Update(r, ctx)
	}

	// If I'm here, I need to poke the backend and fill the cache
	rw := newResponseWriter(r, w, cacheKey)

	rw.Header().Set(GeneratorHeaderKey, UpstreamContentHeaderValue)
	m.next.ServeHTTP(rw, r)
	// the request was served from the upstream.
	// store the response into the cache
	rw.Done(conf.TTL)
}
