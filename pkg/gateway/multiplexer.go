package gateway

import (
	"net/http"

	"github.com/ferama/crauti/pkg/utils"
)

type multiplexer struct {
	muxes      map[string]*http.ServeMux
	defaultMux *http.ServeMux
}

func newMultiplexer() *multiplexer {
	m := &multiplexer{
		defaultMux: http.NewServeMux(),
		muxes:      make(map[string]*http.ServeMux),
	}
	return m
}

func (m *multiplexer) getOrCreate(host string) *http.ServeMux {
	if host == "" {
		return m.defaultMux
	}
	if _, exist := m.muxes[host]; !exist {
		m.muxes[host] = http.NewServeMux()
	}
	return m.muxes[host]
}

func (m *multiplexer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestHost, err := utils.GetRequestHost(r)
	if err != nil {
		log.Fatal().Err(err)
	}
	mux := m.defaultMux
	if _, exist := m.muxes[requestHost]; exist {
		mux = m.muxes[requestHost]
	}
	mux.ServeHTTP(w, r)
}
