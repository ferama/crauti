package proxy

import "testing"

func Test1(t *testing.T) {
	type test struct {
		pattern   string
		transform string
		tests     map[string]string
	}

	tests := []test{
		{
			pattern:   "/bar/(.*)",
			transform: "/baz/$1",
			tests: map[string]string{
				"/bar/test1": "/baz/test1",
			},
		},
		{
			pattern:   `.*/t\?a=(.+)&b=(.+)`,
			transform: `/t/$1/$2`,
			tests: map[string]string{
				"/t?a=1&b=2":    "/t/1/2",
				"/pp/t?a=1&b=2": "/t/1/2",
			},
		},
	}

	for _, v := range tests {
		pm := newRewriter(v.pattern, v.transform)
		for input, expected := range v.tests {
			transformed := pm.rewrite(input)
			if transformed != expected {
				t.Fatalf("expected: %s, got %s", expected, transformed)
			}
		}

	}
}
