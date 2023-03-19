package cache

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ferama/crauti/pkg/cache"
)

type responseWriter struct {
	r   *http.Request
	w   http.ResponseWriter
	buf bytes.Buffer

	statusCode int

	key string
}

func newResponseWriter(
	r *http.Request,
	w http.ResponseWriter,
	key string) *responseWriter {

	rw := &responseWriter{
		r:   r,
		w:   w,
		key: key,
	}
	return rw
}

func (rw *responseWriter) Header() http.Header {
	return rw.w.Header()
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.w.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	rw.buf.Write(data)
	return rw.w.Write(data)
}

func (rw *responseWriter) Done(cacheTTL time.Duration) {
	// build headers cache content. The idea here is to store
	// all the headers sent from backend to send them back to the client
	// when the request hit the cache
	headers := ""
	for k, v := range rw.Header() {
		if headers != "" {
			headers = fmt.Sprintf("%s\r\n%s: %s", headers, k, strings.Join(v, ","))
		} else {
			headers = fmt.Sprintf("%s: %s", k, strings.Join(v, ","))
		}
	}
	// do not cache empty responses if they are not OPTIONS request
	// this fix an issue with the frontend
	if len(rw.buf.Bytes()) > 0 || rw.r.Method == http.MethodOptions {
		cache.Instance().Set(buildRedisKey(headersKeyHead, rw.key), []byte(headers), cacheTTL)
		cache.Instance().Set(buildRedisKey(statusKeyHead, rw.key), rw.statusCode, cacheTTL)
		cache.Instance().Set(buildRedisKey(bodyKeyHead, rw.key), rw.buf.Bytes(), cacheTTL)
	}
}
