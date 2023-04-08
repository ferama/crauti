package gateway

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/ferama/crauti/pkg/chaincontext"
	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/logger"
	"github.com/ferama/crauti/pkg/middleware"
	"github.com/ferama/crauti/pkg/middleware/bodylimit"
	"github.com/ferama/crauti/pkg/middleware/cache"
	"github.com/ferama/crauti/pkg/middleware/collector"
	"github.com/ferama/crauti/pkg/middleware/cors"
	"github.com/ferama/crauti/pkg/middleware/proxy"
	"github.com/ferama/crauti/pkg/middleware/redirect"
	"github.com/ferama/crauti/pkg/middleware/timeout"
	"github.com/ferama/crauti/pkg/utils"
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

type Gateway struct {
	server *server

	updateChan chan *runtimeUpdates
	updateMU   sync.Mutex
}

func NewGateway(httpListenAddr string, httpsListenAddress string) *Gateway {
	s := &Gateway{
		updateChan: make(chan *runtimeUpdates),
	}
	s.server = newServer(httpListenAddr, httpsListenAddress, s.updateChan)

	return s
}

func (s *Gateway) buildRootHandler() http.Handler {
	var chain http.Handler

	chain = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	chain = (&collector.EmitterMiddleware{}).Init(chain)

	next := chain
	chain = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, utils.BodyResponse404)
		next.ServeHTTP(w, r)
	})

	chain = (&collector.CollectorMiddleware{}).Init(chain)

	chain = s.addChainContext(conf.MountPoint{}, chain)
	return chain
}

// get and add a ChainContext instance to the request context
func (s *Gateway) addChainContext(mp conf.MountPoint, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cc := contextPool.Get().(*chaincontext.ChainContext)
		defer contextPool.Put(cc)
		cc.Reset(&mp, cache.CacheStatusMiss)

		rcc := *cc
		r = rcc.Update(r, rcc)
		next.ServeHTTP(w, r)
	})
}

func (s *Gateway) buildChain(mp conf.MountPoint) http.Handler {
	mwares := make([]middleware.Middleware, 0)

	mwares = append(mwares,
		// http -> https
		&redirect.RedirectMiddleware{},
		// collect metrics and logs
		&collector.CollectorMiddleware{},
		// add timetout to context
		&timeout.TimeoutMiddleware{},
		// checks for unwanted large bodies
		&bodylimit.BodyLimiterMiddleware{},
		// add cors headers
		&cors.CorsMiddleware{},
		// respond with cache if we can
		&cache.CacheMiddleware{},
		// poke the backend if needed
		&proxy.ReverseProxyMiddleware{},
	)

	// middelwares are executed in reverse order. the root here is the latest
	// I'm using the for to reverse loop through the mwares slice in order to
	// better read the flow
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	var chain http.Handler
	for i := len(mwares) - 1; i >= 0; i-- {
		if i == len(mwares)-1 {
			chain = root
		}
		chain = mwares[i].Init(chain)
	}
	// setup chain context. This is the first executed
	// middleware (remember the reverse order rule)
	chain = s.addChainContext(mp, chain)
	return chain
}

func (s *Gateway) UpdateHandlers() {
	s.updateMU.Lock()

	collector.MetricsInstance().UnregisterAll()

	mux := newMultiplexer()

	hasRootHandlerDeafault := false
	hasRootHandler := make(map[string]bool)
	hosts := make(map[string]bool)

	log.Print(strings.Repeat("=", 80))

	for _, i := range conf.ConfInst.MountPoints {
		matchHost := i.Middlewares.MatchHost
		if matchHost != "" {
			if _, exists := hasRootHandler[matchHost]; !exists {
				hasRootHandler[matchHost] = false
			}
		}

		log.Debug().
			Str("mountPath", i.Path).
			Str("matchHost", matchHost).
			Msg("registering mount path")

		if matchHost != "" {
			hosts[matchHost] = true
		}

		if i.Path == "/" {
			if matchHost == "" {
				hasRootHandlerDeafault = true
			} else {
				hasRootHandler[matchHost] = true
			}
		}
		// setup metrics
		if i.Path != "" {
			collector.MetricsInstance().RegisterMountPath(i.Path, i.Upstream, matchHost)
		}
		chain := s.buildChain(i)

		mux.getOrCreate(matchHost).Handle(i.Path, chain)
	}

	// if a root path (the / mountPoint) handler was not defined in mountPoints
	// define a custom one here. The root handler, will respond to request for
	// not found resources.

	if !hasRootHandlerDeafault {
		mux.defaultMux.Handle("/", s.buildRootHandler())
	}
	for matchHost, has := range hasRootHandler {
		if !has {
			mux.getOrCreate(matchHost).Handle("/", s.buildRootHandler())
		}
	}

	go func() {
		s.server.stop()
		domains := make([]string, len(hosts))
		for k := range hosts {
			domains = append(domains, k)
		}
		ru := &runtimeUpdates{
			mux:     mux,
			domains: domains,
		}
		s.updateChan <- ru
		s.updateMU.Unlock()
	}()
}

func (s *Gateway) Start() error {
	return s.server.run()
}

func (s *Gateway) Stop() {
	s.server.stop()
}
