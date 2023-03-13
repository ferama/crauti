package kube

import (
	"fmt"
	"sync"
	"time"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/gateway"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type svcUpdater struct {
	server *gateway.Server

	// this field contains
	data map[string]corev1.Service

	mu sync.Mutex
}

func newSvcUpdater(server *gateway.Server) *svcUpdater {
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
		mp := []conf.MountPoint{}

		for _, svc := range s.data {

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
