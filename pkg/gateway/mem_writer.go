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
	rw.header = make(http.Header)
}

func (rw *responseWriter) Header() http.Header {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	return rw.header
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	buf := &bytes.Buffer{}
	rw.bodyBuf = append(rw.bodyBuf, buf)
	return buf.Write(data)
}

func (rw *responseWriter) Data() map[string]any {
	data := make(map[string]any)

	for _, buf := range rw.bodyBuf {
		override := make(map[string]any)
		err := json.Unmarshal(buf.Bytes(), &override)
		if err != nil {
			log.Error().Err(err).Send()
		} else {
			data = mergeMaps(data, override)
		}
	}
	return data
}

func mergeMaps(a, b map[string]any) map[string]any {
	out := make(map[string]any, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]any); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]any); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
