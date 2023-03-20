package collector

import (
	"net/http"
)

type responseWriter struct {
	r *http.Request
	w http.ResponseWriter

	wroteHeader  bool
	statusCode   int
	bytesWritten int
}

func (rw *responseWriter) Status() int {
	return rw.statusCode
}

func (rw *responseWriter) BytesWritten() int {
	return rw.bytesWritten
}

func (rw *responseWriter) Header() http.Header {
	return rw.w.Header()
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	if !rw.wroteHeader {
		rw.statusCode = statusCode
		rw.wroteHeader = true
		rw.WriteHeader(statusCode)
	}
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	n, err := rw.w.Write(data)
	rw.bytesWritten += n
	return n, err
}
