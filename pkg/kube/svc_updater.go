package kube

import (
	"fmt"
	"sync"
	"time"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/gateway"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type svcUpdater struct {
	server *gateway.Server

	// this field contains
	services map[string]corev1.Service

	shouldResync bool

	mu sync.Mutex
}

func newSvcUpdater(server *gateway.Server) *svcUpdater {
	s := &svcUpdater{
		server:       server,
		services:     make(map[string]corev1.Service),
		shouldResync: false,
	}
	go s.synch()

	return s
}

func (s *svcUpdater) synch() {
	parser := new(annotationParser)
	for {
		time.Sleep(10 * time.Second)

		s.mu.Lock()
		if !s.shouldResync {
			s.mu.Unlock()
			continue
		}
		s.shouldResync = false
		mp := []conf.MountPoint{}

		log.Print("== conf update triggered from kube resync ==")
		for _, svc := range s.services {
			annotatedConfig := parser.parse(svc)
			if !annotatedConfig.Enabled {
				continue
			}

			port := annotatedConfig.UpstreamHttpPort
			if port == 0 {
				if len(svc.Spec.Ports) == 1 {
					port = svc.Spec.Ports[0].Port
				} else {
					klog.Error("missing port configuration")
					continue
				}
			}

			for _, item := range annotatedConfig.MountPoints {
				url := fmt.Sprintf("http://%s.%s:%d%s",
					svc.Name, svc.Namespace, port, item.Source)

				mp = append(mp, conf.MountPoint{
					Upstream: url,
					Path:     item.Destination,
					// TODO: needs structure merge?
					// probably not, because the following call to conf.Update()
					// should merge them correctly. Needs testing
					Middlewares: item.Middlewares,
				})
			}
		}
		s.mu.Unlock()

		// update viper conf
		viper.Set("MountPoints", mp)
		// unmarshal the new conf
		conf.Update()
		// update the gateway instance
		s.server.UpdateHandlers()
	}
}

func (s *svcUpdater) add(key string, service corev1.Service) {
	s.mu.Lock()
	defer s.mu.Unlock()

	parser := new(annotationParser)
	s.shouldResync = true
	// if I already have the serivice and the crauti annotation are
	// exactly the same, no resync is required
	if svc, ok := s.services[key]; ok {
		if parser.crautiAnnotationsEquals(svc, service) {
			s.shouldResync = false
		}
	}
	s.services[key] = service
}

func (s *svcUpdater) delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.shouldResync = true
	delete(s.services, key)
}

func (s *svcUpdater) GetAll() map[string]corev1.Service {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.services
}
