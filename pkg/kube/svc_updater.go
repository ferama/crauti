package kube

import (
	"fmt"
	"sync"
	"time"

	"github.com/ferama/crauti/pkg/gateway"
	"github.com/ferama/crauti/pkg/gateway/server"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type svcUpdater struct {
	server *server.Server

	data map[string]corev1.Service

	mu sync.Mutex
}

func newSvcUpdater(server *server.Server) *svcUpdater {
	s := &svcUpdater{
		server: server,
		data:   make(map[string]corev1.Service),
	}
	go s.synch()

	return s
}

func (s *svcUpdater) synch() {
	parser := new(annotationParser)
	for {
		s.mu.Lock()
		mp := []gateway.MountPoint{}

		for _, svc := range s.data {
			conf := parser.parse(svc)
			if !conf.Enabled {
				continue
			}

			port := conf.UpstreamHttpPort
			if port == 0 {
				if len(svc.Spec.Ports) == 1 {
					port = svc.Spec.Ports[0].Port
				} else {
					klog.Error("missing port configuration")
					continue
				}
			}

			for _, item := range conf.MountPoints {
				url := fmt.Sprintf("http://%s.%s:%d%s",
					svc.Name, svc.Namespace, port, item.Source)

				mp = append(mp, gateway.MountPoint{
					Upstream: url,
					Path:     item.Destination,
				})
			}

		}
		s.mu.Unlock()
		s.server.UpdateHandlers(mp)

		time.Sleep(10 * time.Second)
	}
}

func (s *svcUpdater) add(key string, service corev1.Service) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = service
}

func (s *svcUpdater) delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
}

func (s *svcUpdater) GetAll() map[string]corev1.Service {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.data
}
