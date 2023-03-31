package gateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ferama/crauti/pkg/chaincontext"
	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/logger"
	"github.com/ferama/crauti/pkg/middleware/bodylimit"
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
			ReadTimeout:  conf.ConfInst.Gateway.ReadTimeout,
			WriteTimeout: conf.ConfInst.Gateway.WriteTimeout,
			IdleTimeout:  conf.ConfInst.Gateway.IdleTimeout,
			Addr:         listenAddr,
		},
	}
	s.UpdateHandlers()

	return s
}

func (s *Server) buildRootHandler() http.Handler {
	var chain http.Handler

	chain = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	chain = collector.NewEmitterrMiddleware(chain)

	next := chain
	chain = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "crauti: 404 not found\n")
		next.ServeHTTP(w, r)
	})

	chain = collector.NewCollectorMiddleware(chain)
	return chain
}

func (s *Server) buildChain(mp conf.MountPoint) http.Handler {
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	var chain http.Handler
	chain = root

	chain = collector.NewEmitterrMiddleware(chain)
	// this need to run just before the logEmitter one (remember the reverse order of run)
	chain = timeout.NewTimeoutHandlerMiddleware(chain)
	// Middlewares are executed in reverse order: the last one
	// is exectuted first
	chain = proxy.NewReverseProxyMiddleware(chain)
	// install the cache middleware
	chain = cache.NewCacheMiddleware(chain)
	// install the cors middleware
	chain = cors.NewCorsMiddleware(chain)

	chain = bodylimit.NewBodyLimiterMiddleware(chain)
	chain = timeout.NewTimeoutMiddleware(chain)
	// should be the first middleware to be able to measure
	// stuff like time, bytes etc
	chain = collector.NewCollectorMiddleware(chain)

	// setup chain context
	next := chain
	chain = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), chaincontext.ChainContextKey, chaincontext.ChainContext{
			Conf: mp,
		})
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
	return chain
}

func (s *Server) UpdateHandlers() {
	collector.MetricsInstance().UnregisterAll()

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
		// setup metrics
		if i.Path != "" {
			collector.MetricsInstance().RegisterMountPath(i.Path, i.Upstream)
		}
		chain := s.buildChain(i)
		mux.Handle(i.Path, chain)
	}

	// if a root path (the / mountPoint) handler was not defined in mountPoints
	// define a custom one here. The root handler, will respond to request for
	// not found resources.
	if !hasRootHandler {
		mux.Handle("/", s.buildRootHandler())
	}
	s.srv.Handler = mux
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}
