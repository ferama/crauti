package cache

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ferama/crauti/pkg/redis"
)

type responseWriter struct {
	r *http.Request
	w http.ResponseWriter

	bodyBuf bytes.Buffer

	statusCode int

	cacheKey string
}

func newResponseWriter(
	r *http.Request,
	w http.ResponseWriter,
	cacheKey string) *responseWriter {

	rw := &responseWriter{
		r:        r,
		w:        w,
		cacheKey: cacheKey,
	}
	return rw
}

func (rw *responseWriter) Header() http.Header {
	return rw.w.Header()
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	// store the status code to be able to cache it laters
	rw.statusCode = statusCode
	rw.w.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	rw.bodyBuf.Write(data)
	return rw.w.Write(data)
}

func (rw *responseWriter) Done(cacheTTL time.Duration) {
	// build headers cache content. The idea here is to store
	// all the headers sent from backend to send them back to the client
	// when the request hit the cache
	headers := ""
	for k, v := range rw.Header() {
		if k == "X-Generator" {
			continue
		}
		if headers != "" {
			headers = fmt.Sprintf("%s\r\n%s: %s", headers, k, strings.Join(v, ","))
		} else {
			headers = fmt.Sprintf("%s: %s", k, strings.Join(v, ","))
		}
	}
	// do not cache empty responses if they are not OPTIONS or HEAD request
	if len(rw.bodyBuf.Bytes()) > 0 ||
		rw.r.Method == http.MethodOptions ||
		rw.r.Method == http.MethodHead {
		// headers
		redis.CacheInstance().Set(buildRedisKey(headersKeyHead, rw.cacheKey), []byte(headers), cacheTTL)
		// status
		redis.CacheInstance().Set(buildRedisKey(statusKeyHead, rw.cacheKey), rw.statusCode, cacheTTL)
		// body
		redis.CacheInstance().Set(buildRedisKey(bodyKeyHead, rw.cacheKey), rw.bodyBuf.Bytes(), cacheTTL)
	}
}
