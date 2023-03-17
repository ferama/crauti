package gateway

import (
	"log"
	"net/http"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/middleware"
	"github.com/ferama/crauti/pkg/middleware/cache"
	"github.com/ferama/crauti/pkg/middleware/cors"
)

type Server struct {
	srv *http.Server
}

func NewServer(listenAddr string) *Server {
	s := &Server{
		srv: &http.Server{
			// ReadHeaderTimeout: 5 * time.Second,
			// ReadTimeout:       5 * time.Second,
			// WriteTimeout:      10 * time.Second,
			Addr: listenAddr,
		},
	}
	s.UpdateHandlers()

	return s
}

func (s *Server) UpdateHandlers() {
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := http.NewServeMux()

	defer func() {
		if err := recover(); err != nil {
			log.Println("[ERROR] panic occurred:", err)
			s.srv.Handler = mux
		}
	}()

	for _, i := range conf.ConfInst.MountPoints {

		var chain http.Handler
		chain = root

		// Middlewares are executed in reverse order: the last one
		// is exectuted first
		chain, _ = middleware.NewReverseProxyMiddleware(chain, i)

		cacheConf := i.Middlewares.Cache
		if cacheConf.IsEnabled() {
			chain = cache.NewCacheMiddleware(
				chain,
				cacheConf,
			)
		}

		chain = middleware.NewTimeoutMiddleware(chain, i.Middlewares.Timeout)

		corsConf := i.Middlewares.Cors
		if corsConf.IsEnabled() {
			// install the cors middleware
			chain = cors.NewCorsMiddleware(chain)
		}

		mux.Handle(i.Path, chain)
	}

	s.srv.Handler = mux
}

func (s *Server) Start() {
	s.srv.ListenAndServe()
}
