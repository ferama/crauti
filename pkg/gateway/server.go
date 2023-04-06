package gateway

import (
	"context"
	"crypto/tls"
	"net/http"
	"sync"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/kube/certcache"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

// brings updates from the gateway when config
// change at runtime
type runtimeUpdates struct {
	// the new multiplexer with updated mountPoints
	mux *multiplexer

	// a list of served domains
	domains []string
}

type server struct {
	// if true, server will listen on https port too
	// either using provided certs or using the Let's Encrypt
	// ones if autoHTTPEnabled
	HTTPSEnabled bool
	// if true, generates certificates using the Let's Encrypt
	// You also need to set HTTPSEnabled = true
	// If you don't do it, autoHTTPSEnabled will have no effect
	autoHTTPSEnabled bool

	https *http.Server
	http  *http.Server

	httpListenAddr  string
	httpsListenAddr string

	updateChan chan *runtimeUpdates

	mu sync.Mutex
}

func newServer(httpListenAddr string, httpsListenAddress string, update chan *runtimeUpdates) *server {

	HTTPEnabled := conf.ConfInst.Gateway.HTTPSEnabled
	autoHTTPSEnabled := conf.ConfInst.Gateway.AutoHTTPSEnabled
	if autoHTTPSEnabled {
		HTTPEnabled = true
		autoHTTPSEnabled = true
	}

	s := &server{
		HTTPSEnabled:     HTTPEnabled,
		autoHTTPSEnabled: autoHTTPSEnabled,
		https:            &http.Server{},
		http:             &http.Server{},
		httpListenAddr:   httpListenAddr,
		httpsListenAddr:  httpsListenAddress,
		updateChan:       update,
	}

	return s
}

func (s *server) setupServers(updates *runtimeUpdates) {
	var handler http.Handler
	handler = updates.mux

	if s.HTTPSEnabled {
		s.https = &http.Server{
			ReadTimeout:  conf.ConfInst.Gateway.ReadTimeout,
			WriteTimeout: conf.ConfInst.Gateway.WriteTimeout,
			IdleTimeout:  conf.ConfInst.Gateway.IdleTimeout,
			Handler:      handler,
			Addr:         s.httpsListenAddr,
		}
	}

	if s.HTTPSEnabled && s.autoHTTPSEnabled {

		certManager := autocert.Manager{
			Client: &acme.Client{
				DirectoryURL: "https://acme-staging-v02.api.letsencrypt.org/directory",
				// DirectoryURL: "https://acme-v02.api.letsencrypt.org/directory",
			},
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(updates.domains...),
		}
		if conf.ConfInst.Gateway.AutoHTTPSUseLocalDir {
			dir := conf.ConfInst.Gateway.AutoHTTPSLocalDir
			certManager.Cache = autocert.DirCache(dir)
		} else {
			certManager.Cache = certcache.NewSecretCache("")
		}

		s.https.TLSConfig = certManager.TLSConfig()
		handler = certManager.HTTPHandler(handler)
	}

	s.http = &http.Server{
		ReadTimeout:  conf.ConfInst.Gateway.ReadTimeout,
		WriteTimeout: conf.ConfInst.Gateway.WriteTimeout,
		IdleTimeout:  conf.ConfInst.Gateway.IdleTimeout,
		Addr:         s.httpListenAddr,
		Handler:      handler,
	}
}

func (s *server) buildTLSConfig() *tls.Config {
	cfg := &tls.Config{}

	for _, v := range conf.ConfInst.Gateway.KeyPairs {
		crt, err := tls.LoadX509KeyPair(v.FullChain, v.PrivateKey)
		if err == nil {
			cfg.Certificates = append(cfg.Certificates, crt)
		}
	}

	return cfg
}

func (s *server) run() error {
	var wg sync.WaitGroup

	for {
		// ensures that both http and https servers are down
		wg.Wait()

		updates := <-s.updateChan

		s.mu.Lock()
		s.setupServers(updates)
		s.mu.Unlock()

		log.Print("starting new server...")

		wg.Add(1)
		go func() {
			log.Printf("http - %s", s.http.ListenAndServe())
			wg.Done()
		}()

		if s.HTTPSEnabled {
			if !s.autoHTTPSEnabled {
				s.https.TLSConfig = s.buildTLSConfig()
			}
			wg.Add(1)
			log.Printf("https - %s", s.https.ListenAndServeTLS("", ""))
			wg.Done()
		}
	}
}

func (s *server) stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.HTTPSEnabled {
		s.https.Shutdown(context.Background())
	}
	s.http.Shutdown(context.Background())
}
