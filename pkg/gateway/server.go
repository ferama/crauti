package gateway

import (
	"fmt"
	"net/http"
	"sync"

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

var (
	log         *zerolog.Logger
	contextPool sync.Pool
)

func init() {
	contextPool = sync.Pool{
		New: func() any {
			return chaincontext.NewChainContext()
		},
	}
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

	chain = s.addChainContext(conf.MountPoint{}, chain)
	return chain
}

func (s *Server) addChainContext(mp conf.MountPoint, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cc := contextPool.Get().(*chaincontext.ChainContext)
		defer contextPool.Put(cc)
		cc.Reset(&mp, cache.CacheStatusMiss)

		rcc := *cc
		r = rcc.Update(r, rcc)
		next.ServeHTTP(w, r)
	})
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
	chain = s.addChainContext(mp, chain)
	return chain
}

func (s *Server) UpdateHandlers() {
	collector.MetricsInstance().UnregisterAll()

	multiplexer := newMultiplexer()

	if !conf.ConfInst.Debug {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("== %s", err)
				s.srv.Handler = multiplexer.defaultMux
			}
		}()
	}

	hasRootHandler := false
	log.Print("===============================================================")
	for _, i := range conf.ConfInst.MountPoints {
		matchHost := i.Middlewares.Proxy.MatchHost
		log.Debug().
			Str("mountPath", i.Path).
			Str("matchHost", matchHost).
			Msg("registering mount path")

		if i.Path == "/" {
			hasRootHandler = true
		}
		// setup metrics
		if i.Path != "" {
			collector.MetricsInstance().RegisterMountPath(i.Path, i.Upstream, matchHost)
		}
		chain := s.buildChain(i)

		multiplexer.getOrCreate(matchHost).Handle(i.Path, chain)
	}

	// if a root path (the / mountPoint) handler was not defined in mountPoints
	// define a custom one here. The root handler, will respond to request for
	// not found resources.
	if !hasRootHandler {
		multiplexer.defaultMux.Handle("/", s.buildRootHandler())
	}
	s.srv.Handler = multiplexer

}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}
