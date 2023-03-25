package gateway

import (
	"fmt"
	"net/http"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/logger"
	"github.com/ferama/crauti/pkg/middleware/cache"
	"github.com/ferama/crauti/pkg/middleware/collector"
	"github.com/ferama/crauti/pkg/middleware/cors"
	"github.com/ferama/crauti/pkg/middleware/proxy"
	"github.com/ferama/crauti/pkg/middleware/timeout"
	"github.com/rs/zerolog"
)

var log *zerolog.Logger

func init() {
	log = logger.GetLogger("gateway")
}

type Server struct {
	srv *http.Server
}

func NewServer(listenAddr string) *Server {
	s := &Server{
		srv: &http.Server{
			// ReadHeaderTimeout: 5 * time.Second,
			// WriteTimeout:      10 * time.Second,
			// IdleTimeout:       120 * time.Second,
			Addr: listenAddr,
		},
	}
	s.UpdateHandlers()

	return s
}

func (s *Server) setupRootHandler(mux *http.ServeMux) {
	var chain http.Handler

	chain = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	chain = collector.NewEmitterrMiddleware(chain, "")

	next := chain
	chain = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "crauti: 404 not found\n")
		next.ServeHTTP(w, r)
	})

	chain = collector.NewCollectorMiddleware(chain)
	mux.Handle("/", chain)
}

func (s *Server) UpdateHandlers() {
	collector.MetricsInstance().UnregisterAll()

	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := http.NewServeMux()

	defer func() {
		if err := recover(); err != nil {
			log.Printf("%s", err)
			s.srv.Handler = mux
		}
	}()

	hasRootHandler := false
	for _, i := range conf.ConfInst.MountPoints {

		if i.Path == "/" {
			hasRootHandler = true
		}

		var chain http.Handler
		chain = root

		chain = collector.NewEmitterrMiddleware(chain, i.Path)
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

	// if a root path (the / mountPoint) handler was not defined in mountPoints
	// define a custom one here. The root handler, will respond to request for
	// not found resources.
	if !hasRootHandler {
		s.setupRootHandler(mux)
	}
	s.srv.Handler = mux
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}
