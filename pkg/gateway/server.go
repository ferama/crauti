package gateway

import (
	"net/http"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/middleware/cors"
	"github.com/ferama/crauti/pkg/middleware/reverseproxy"
)

type Server struct {
	srv *http.Server

	mountPoints []conf.MountPoint
}

func NewServer(listenAddr string) *Server {
	s := &Server{
		srv: &http.Server{
			// ReadHeaderTimeout: 5 * time.Second,
			// ReadTimeout:       5 * time.Second,
			// WriteTimeout:      10 * time.Second,
			Addr: listenAddr,
		},
		mountPoints: make([]conf.MountPoint, 0),
	}
	s.UpdateHandlers(nil)

	return s
}

func (s *Server) UpdateHandlers(mountPoints []conf.MountPoint) {
	s.mountPoints = mountPoints

	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := http.NewServeMux()
	for _, i := range mountPoints {

		var chain http.Handler
		chain = root

		// Middlewares are executed in reverse order: the last one
		// is exectuted first
		chain, _ = reverseproxy.NewReverseProxy(chain, i)

		// install the cors middleware
		chain = cors.NewCors(chain)
		mux.Handle(i.Path, chain)
	}

	s.srv.Handler = mux
}

func (s *Server) Start() {
	s.srv.ListenAndServe()
}
