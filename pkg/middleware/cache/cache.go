package cache

import (
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	red    = "\033[0;31m"
	green  = "\033[0;32m"
	yellow = "\033[0;33m"
	blue   = "\033[0;34m"
	reset  = "\033[0m"
)

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

type cacheMiddleware struct {
	next http.Handler

	// cache this httpMethods only
	httpMethods []string
	// this headers will be considered to build the cache key
	keyHeaders []string

	// cache time to live
	cacheTTL *time.Duration

	mu      sync.Mutex
	lockmap map[string]*sync.Mutex
}

func NewCacheMiddleware(
	next http.Handler,
	httpMethods []string,
	keyHeaders []string,
	cacheTTL *time.Duration,
) (http.Handler, error) {

	kh := []string{}
	c := cases.Title(language.English)
	for _, h := range keyHeaders {
		kh = append(kh, c.String(h))
	}

	cm := &cacheMiddleware{
		next:        next,
		keyHeaders:  kh,
		httpMethods: httpMethods,
		cacheTTL:    cacheTTL,
		lockmap:     make(map[string]*sync.Mutex),
	}

	return cm, nil
}

func (m *cacheMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// if the request should not be cached because the http
	// method needs to be ignored or because it is disabled,
	// directly serve it ignoring the cache
	if !contains(m.httpMethods, r.Method) {
		log.Printf("%s[BYP]%s %s%s ", blue, reset, r.Method, r.URL)
		m.next.ServeHTTP(w, r)
		return
	}
}
