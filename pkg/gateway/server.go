package gateway

import (
	"context"
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
	"github.com/ferama/crauti/pkg/middleware/timeout"
	"github.com/ferama/crauti/pkg/utils"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
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
	autoHttpsEnabled bool

	https *http.Server
	http  *http.Server

	httpListenAddr  string
	httpsListenAddr string

	updateChan chan *multiplexer
	updateMU   sync.Mutex

	hosts map[string]bool
}

func NewServer(httpListenAddr string, httpsListenAddress string) *Server {
	s := &Server{
		autoHttpsEnabled: false,
		https:            &http.Server{},
		http:             &http.Server{},
		httpListenAddr:   httpListenAddr,
		httpsListenAddr:  httpsListenAddress,
		updateChan:       make(chan *multiplexer),
		hosts:            make(map[string]bool),
	}

	return s
}

func (s *Server) buildRootHandler() http.Handler {
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
	mwares := make([]middleware.Middleware, 0)

	mwares = append(mwares,
		// collect metrics and logs
		&collector.CollectorMiddleware{},
		// add timetout to context
		&timeout.TimeoutMiddleware{},
		// checks for unwanted large bodies
		&bodylimit.BodyLimiter{},
		// add cors headers
		&cors.CorsMiddleware{},
		// respond with cache if we can
		&cache.CacheMiddleware{},
		// poke the backend if needed
		&proxy.ReverseProxyMiddleware{},
		// respond with a bad gateway message on timeout
		&timeout.TimeoutHandlerMiddleware{},
		// emits collected logs and metrics
		&collector.EmitterMiddleware{},
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

func (s *Server) UpdateHandlers() {
	s.updateMU.Lock()

	collector.MetricsInstance().UnregisterAll()

	mux := newMultiplexer()

	hasRootHandlerDeafault := false
	hasRootHandler := make(map[string]bool)

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
			s.hosts[matchHost] = true
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
		s.Stop()
		s.updateChan <- mux
		s.updateMU.Unlock()
	}()
}

func (s *Server) Start() error {
	var wg sync.WaitGroup

	for {
		wg.Wait()

		mux := <-s.updateChan

		s.https = &http.Server{
			ReadTimeout:  conf.ConfInst.Gateway.ReadTimeout,
			WriteTimeout: conf.ConfInst.Gateway.WriteTimeout,
			IdleTimeout:  conf.ConfInst.Gateway.IdleTimeout,
			Handler:      mux,
			Addr:         s.httpsListenAddr,
		}

		domains := make([]string, len(s.hosts))
		for k := range s.hosts {
			domains = append(domains, k)
		}
		certManager := autocert.Manager{
			Client: &acme.Client{
				DirectoryURL: "https://acme-staging-v02.api.letsencrypt.org/directory",
				// DirectoryURL: "https://acme-v02.api.letsencrypt.org/directory",
			},
			Cache:      autocert.DirCache("./certs-cache"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(domains...),
		}
		s.https.TLSConfig = certManager.TLSConfig()

		// TODO: needs a custom fallback handler that redirect to https
		// mountPoints with matchHost only and fallback to http the rest
		var handler http.Handler
		handler = mux
		if s.autoHttpsEnabled {
			handler = certManager.HTTPHandler(handler)
		}
		s.http = &http.Server{
			ReadTimeout:  conf.ConfInst.Gateway.ReadTimeout,
			WriteTimeout: conf.ConfInst.Gateway.WriteTimeout,
			IdleTimeout:  conf.ConfInst.Gateway.IdleTimeout,
			Addr:         s.httpListenAddr,
			Handler:      handler,
		}

		log.Print("starting new server...")

		wg.Add(1)
		go func() {
			log.Printf("http - %s", s.http.ListenAndServe())
			wg.Done()
		}()

		if s.autoHttpsEnabled {
			wg.Add(1)
			log.Printf("https - %s", s.https.ListenAndServeTLS("", ""))
			wg.Done()
		}
	}
}

func (s *Server) Stop() {
	if s.autoHttpsEnabled {
		s.https.Shutdown(context.Background())
	}
	s.http.Shutdown(context.Background())
}
