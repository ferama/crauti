package gateway

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
)

type responseWriter struct {
	r *http.Request

	bodyBuf []*bytes.Buffer
	mu      sync.Mutex

	header http.Header
}

func (rw *responseWriter) Reset(r *http.Request, w http.ResponseWriter) {
	rw.r = r
	rw.header = make(http.Header)
	rw.bodyBuf = make([]*bytes.Buffer, 0)
}

func (rw *responseWriter) Header() http.Header {
	return rw.header
}

func (rw *responseWriter) WriteHeader(statusCode int) {
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	buf := &bytes.Buffer{}
	rw.bodyBuf = append(rw.bodyBuf, buf)
	return buf.Write(data)
}

func (rw *responseWriter) Data() map[string]interface{} {
	data := make(map[string]interface{})

	for _, buf := range rw.bodyBuf {
		override := make(map[string]interface{})
		err := json.Unmarshal(buf.Bytes(), &override)
		if err != nil {
			// panic(err)
		} else {
			data = mergeMaps(data, override)
		}
	}
	return data
}

func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
