package kube

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestParser(t *testing.T) {
	annotations := map[string]string{
		crautiAnnotationConfKey: `{
				"enabled": true,
				"upstreamHttpPort": 8080,
				"mountPoints": [
					{"source": "/", "path": "/api/config"}
				]
		}`,
	}
	svc := new(corev1.Service)
	svc.Annotations = annotations

	parser := new(annotationParser)
	conf := parser.parse(*svc)

	if conf.Enabled != true {
		t.Fatal("true expected")
	}

	if conf.UpstreamHttpPort != 8080 {
		t.Fatal("8080 expected")
	}

	if conf.MountPoints[0].Path != "/api/config" {
		t.Fatal("/api/config expected")
	}

	annotations = map[string]string{
		crautiAnnotationConfKey: `
enabled: true
upstreamHttpPort: 8080
mountPoints:
  - source: "/"
    path: "/api/config"
`}
	svc.Annotations = annotations
	conf = parser.parse(*svc)

	if conf.Enabled != true {
		t.Fatal("true expected")
	}

	if conf.UpstreamHttpPort != 8080 {
		t.Fatal("8080 expected")
	}

	if conf.MountPoints[0].Path != "/api/config" {
		t.Fatal("/api/config expected")
	}
}
