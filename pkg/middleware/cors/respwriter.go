package cors

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

type responseWriter struct {
	r *http.Request
	w http.ResponseWriter
}

func (rw *responseWriter) Reset(r *http.Request, w http.ResponseWriter) {
	rw.r = r
	rw.w = w
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.w.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return h.Hijack()
}

func (rw *responseWriter) Header() http.Header {
	return rw.w.Header()
}

// test it with something like
//
//	curl -v -H "Origin: localhost" http://localhost:8080/get
func (rw *responseWriter) WriteHeader(statusCode int) {
	// remove cors headers they could be cached. we will
	// add them back soon if the Origin header is present
	rw.w.Header().Del("Access-Control-Allow-Origin")
	rw.w.Header().Del("Access-Control-Allow-Methods")
	rw.w.Header().Del("Access-Control-Allow-Headers")
	rw.w.Header().Del("Access-Control-Allow-Credentials")
	// always put cors headers. If do not do so, we could have cached headers
	// from a basic curl request that doesn't include the cors
	origin := rw.r.Header.Get("Origin")
	if origin != "" {
		rw.w.Header().Set("Access-Control-Allow-Origin", rw.r.Header.Get("Origin"))
		rw.w.Header().Set("Access-Control-Allow-Methods", rw.r.Method)
		rw.w.Header().Set("Access-Control-Allow-Headers", "*")
	}
	rw.w.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	return rw.w.Write(data)
}
