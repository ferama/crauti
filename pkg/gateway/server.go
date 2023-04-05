package gateway

import (
	"context"
	"net/http"
	"sync"

	"github.com/ferama/crauti/pkg/conf"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

type runtimeUpdates struct {
	mux     *multiplexer
	domains []string
}

type server struct {
	autoHttpsEnabled bool

	https *http.Server
	http  *http.Server

	httpListenAddr  string
	httpsListenAddr string

	updateChan chan *runtimeUpdates
}

func newServer(httpListenAddr string, httpsListenAddress string, update chan *runtimeUpdates) *server {
	s := &server{
		autoHttpsEnabled: false,
		https:            &http.Server{},
		http:             &http.Server{},
		httpListenAddr:   httpListenAddr,
		httpsListenAddr:  httpsListenAddress,
		updateChan:       update,
	}

	return s
}

func (s *server) Start() error {
	var wg sync.WaitGroup

	for {
		wg.Wait()

		updates := <-s.updateChan

		s.https = &http.Server{
			ReadTimeout:  conf.ConfInst.Gateway.ReadTimeout,
			WriteTimeout: conf.ConfInst.Gateway.WriteTimeout,
			IdleTimeout:  conf.ConfInst.Gateway.IdleTimeout,
			Handler:      updates.mux,
			Addr:         s.httpsListenAddr,
		}

		certManager := autocert.Manager{
			Client: &acme.Client{
				DirectoryURL: "https://acme-staging-v02.api.letsencrypt.org/directory",
				// DirectoryURL: "https://acme-v02.api.letsencrypt.org/directory",
			},
			Cache:      autocert.DirCache("./certs-cache"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(updates.domains...),
		}
		s.https.TLSConfig = certManager.TLSConfig()

		// TODO: needs a custom fallback handler that redirect to https
		// mountPoints with matchHost only and fallback to http the rest
		var handler http.Handler
		handler = updates.mux
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

func (s *server) Stop() {
	if s.autoHttpsEnabled {
		s.https.Shutdown(context.Background())
	}
	s.http.Shutdown(context.Background())
}
