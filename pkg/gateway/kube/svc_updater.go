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
)

type svcUpdater struct {
	server *gateway.Gateway

	// this field contains
	services map[string]corev1.Service

	shouldResync bool

	mu sync.Mutex
}

func newSvcUpdater(server *gateway.Gateway) *svcUpdater {
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
					log.Error().Msg("missing port configuration")
					continue
				}
			}

			for _, item := range annotatedConfig.MountPoints {
				url := fmt.Sprintf("http://%s.%s:%d%s",
					svc.Name, svc.Namespace, port, item.Source)

				mp = append(mp, conf.MountPoint{
					Upstream:  url,
					Path:      item.Path,
					MatchHost: item.MatchHost,
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
		s.server.Update()
	}
}

func (s *svcUpdater) onAdd(obj interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	svc := obj.(*corev1.Service)
	key := fmt.Sprintf("%s/%s", svc.Namespace, svc.Name)
	service := *svc.DeepCopy()

	parser := new(annotationParser)
	s.shouldResync = true
	// if I already have the serivice and the crauti annotation are
	// exactly the same, no resync is required
	if svc, ok := s.services[key]; ok {
		if parser.crautiAnnotationsEqual(svc, service) {
			s.shouldResync = false
		}
	}
	s.services[key] = service
}

func (s *svcUpdater) onUpdate(oldObj interface{}, newObj interface{}) {
	s.onAdd(newObj)
}

func (s *svcUpdater) onDelete(obj interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	svc := obj.(*corev1.Service)
	key := fmt.Sprintf("%s/%s", svc.Namespace, svc.Name)

	s.shouldResync = true
	delete(s.services, key)
}

func (s *svcUpdater) GetAll() map[string]corev1.Service {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.services
}
