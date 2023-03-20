package gateway

import (
	"log"
	"net/http"
	"time"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/middleware/cache"
	"github.com/ferama/crauti/pkg/middleware/collector"
	"github.com/ferama/crauti/pkg/middleware/cors"
	"github.com/ferama/crauti/pkg/middleware/proxy"
	"github.com/ferama/crauti/pkg/middleware/timeout"
)

type Server struct {
	srv *http.Server
}

func NewServer(listenAddr string) *Server {
	s := &Server{
		srv: &http.Server{
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       120 * time.Second,
			Addr:              listenAddr,
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

		chain = collector.NewLogEmitterrMiddleware(chain)
		// this need to run just before the logEmitter one (remember the reverse order of run)
		chain = timeout.NewTimeoutHandlerMiddleware(chain)
		// Middlewares are executed in reverse order: the last one
		// is exectuted first
		chain = proxy.NewReverseProxyMiddleware(chain, i)

		cacheConf := i.Middlewares.Cache
		if cacheConf.IsEnabled() {
			chain = cache.NewCacheMiddleware(
				chain,
				cacheConf,
			)
		}

		corsConf := i.Middlewares.Cors
		if corsConf.IsEnabled() {
			// install the cors middleware
			chain = cors.NewCorsMiddleware(chain)
		}

		chain = timeout.NewTimeoutMiddleware(chain, i.Middlewares.Timeout)
		// should be the first middleware to be able to measure
		// stuff like time, bytes etc
		chain = collector.NewCollectorMiddleware(chain)
		mux.Handle(i.Path, chain)
	}

	s.srv.Handler = mux
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}
