package kube

import (
	"github.com/ferama/crauti/pkg/conf"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const crautiAnnotationConfKey = "crauti/conf"

/*
Annotations example:

	annotations:
		crauti/conf: |
			{"enabled": true, "mountPoints": [
				{"source": "/", "path": "/api/test"},
				{"source": "/", "path": "/test2"}
			]}

or you can use yaml instead

	annotations:
		crauti/conf: |
			enabled: true
			# optional: force upstream port
			upstreamHttpPort: 8181
			mountPoints:
				- source: "/"
				  path: "/test1-t1"
				  # overrides global middlewares conf
				  middlewares:
					cors:
					  enabled: true
				- source: "/"
    			  path: "/test2"

*/

type annotationMountPoint struct {
	// struct composition with deault conf MountPoint definition
	conf.MountPoint

	// Custom fields for better user experience while
	// using conf on service annotations
	Source string `yaml:"source"`
	Path   string `yaml:"path"`
}

// this is the service annotation config. It will be mapped
// to crauti configuguration
type crautiAnnotatedConfig struct {
	Enabled          bool                   `yaml:"enabled"`
	UpstreamHttpPort int32                  `yaml:"upstreamHttpPort"`
	MountPoints      []annotationMountPoint `yaml:"mountPoints"`
}

type annotationParser struct{}

func (a *annotationParser) parse(svc corev1.Service) *crautiAnnotatedConfig {

	config := &crautiAnnotatedConfig{}

	for key, value := range svc.Annotations {
		if key == crautiAnnotationConfKey {
			err := yaml.Unmarshal([]byte(value), config)
			if err != nil {
				log.Err(err)
				continue
			}
		}
	}

	return config
}

func (a *annotationParser) crautiAnnotationsEqual(svc1 corev1.Service, svc2 corev1.Service) bool {
	var ann1, ann2 string
	for key, value := range svc1.Annotations {
		if key == crautiAnnotationConfKey {
			ann1 = value
			break
		}
	}
	for key, value := range svc2.Annotations {
		if key == crautiAnnotationConfKey {
			ann2 = value
			break
		}
	}

	return ann1 == ann2
}
