package kube

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog/v2"
)

const crautiAnnotationConfKey = "crauti/conf"

/*
Annotations example:

	annotations:
		crauti/conf: |
			{"enabled": true, "mountPoints": [
				{"source": "/", "destination": "/api/test"},
				{"source": "/", "destination": "/test2"}
			]}

or you can use yaml instead

	annotations:
		crauti/conf: |
			enabled: true
			mountPoints:
			  - source: "/"
    			destination: "/api/test"
			  - source: "/"
    			destination: "/test2"

*/

type annotationMountPoint struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

// this is the service annotation config. It will be mapped
// to crauti configuguration
type crautiAnnotatedConfig struct {
	Enabled          bool                   `json:"enabled"`
	UpstreamHttpPort int32                  `json:"upstreamHttpPort"`
	MountPoints      []annotationMountPoint `json:"mountPoints"`
}

type annotationParser struct{}

func (a *annotationParser) parse(svc corev1.Service) *crautiAnnotatedConfig {

	config := &crautiAnnotatedConfig{}

	for key, value := range svc.Annotations {
		if key == crautiAnnotationConfKey {
			err := yaml.Unmarshal([]byte(value), config)
			if err != nil {
				klog.Error(err)
				continue
			}
		}
	}

	return config
}
