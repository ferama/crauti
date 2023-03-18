package gateway

import (
	"log"
	"net/http"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/middleware/cache"
	"github.com/ferama/crauti/pkg/middleware/cors"
	loggermiddleware "github.com/ferama/crauti/pkg/middleware/logger"
	"github.com/ferama/crauti/pkg/middleware/proxy"
	"github.com/ferama/crauti/pkg/middleware/timeout"
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
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			w.WriteHeader(http.StatusGatewayTimeout)
			w.Write([]byte("bad gateway: connection timeout\n"))
		default:
		}
		loggermiddleware.Printer(w, r)
	})

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
		chain, _ = proxy.NewReverseProxyMiddleware(chain, i)

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
		chain = loggermiddleware.NewLogCollectorMiddleware(chain)
		mux.Handle(i.Path, chain)
	}

	s.srv.Handler = mux
}

func (s *Server) Start() {
	s.srv.ListenAndServe()
}
